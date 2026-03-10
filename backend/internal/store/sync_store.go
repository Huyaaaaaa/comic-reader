package store

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"gorm.io/gorm"
)

type SyncStore struct {
	db *gorm.DB
}

func NewSyncStore(db *gorm.DB) *SyncStore {
	return &SyncStore{db: db}
}

func (store *SyncStore) CreateJob(jobType string, paramsJSON string) (int64, error) {
	result := struct {
		ID int64 `gorm:"column:id"`
	}{}
	err := store.db.Raw(`
INSERT INTO sync_jobs (job_type, status, params_json, progress, created_at, updated_at)
VALUES (?, 'running', ?, 0, datetime('now'), datetime('now'))
RETURNING id
`, jobType, paramsJSON).Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (store *SyncStore) UpdateJob(jobID int64, status string, progress float64, lastError string) error {
	changes := map[string]any{
		"status":     status,
		"progress":   progress,
		"last_error": lastError,
		"updated_at": time.Now(),
	}
	if status == "running" {
		changes["started_at"] = time.Now()
	}
	if status == "completed" || status == "failed" || status == "canceled" {
		changes["finished_at"] = time.Now()
	}
	return store.db.Table("sync_jobs").Where("id = ?", jobID).Updates(changes).Error
}

func (store *SyncStore) SaveCatalogSnapshot(sourceID int64, totalComics int, totalPages int, lastPageCount int) error {
	return store.db.Exec(`
INSERT INTO catalog_snapshots (source_id, total_comics, total_pages, last_page_count, captured_at)
VALUES (?, ?, ?, ?, datetime('now'))
`, sourceID, totalComics, totalPages, lastPageCount).Error
}

func (store *SyncStore) UpsertHeadPage(source *models.SourceSite, page models.RemoteListPage) (int, error) {
	updated := 0
	err := store.db.Transaction(func(tx *gorm.DB) error {
		for index, item := range page.Items {
			if item.ID == 0 {
				continue
			}
			if err := tx.Exec(`
INSERT INTO comics (
  id, title, subtitle, cover_url, rating, rating_count, favorites_remote,
  source_last_seen_at, created_at, updated_at
) VALUES (?, ?, '', ?, ?, ?, ?, datetime('now'), datetime('now'), datetime('now'))
ON CONFLICT(id) DO UPDATE SET
  title = CASE WHEN excluded.title != '' THEN excluded.title ELSE comics.title END,
  cover_url = CASE WHEN excluded.cover_url != '' THEN excluded.cover_url ELSE comics.cover_url END,
  rating = excluded.rating,
  rating_count = excluded.rating_count,
  favorites_remote = excluded.favorites_remote,
  source_last_seen_at = datetime('now'),
  updated_at = datetime('now')
`, item.ID, item.Title, item.CoverURL, item.Rating, item.RatingCount, item.Favorites).Error; err != nil {
				return err
			}

			sortKey := float64((page.CurrentPage-1)*1000 + index + 1)
			if err := tx.Exec(`
INSERT INTO comic_order_index (comic_id, sort_key, source, remote_page, remote_pos, order_updated_at)
VALUES (?, ?, ?, ?, ?, datetime('now'))
ON CONFLICT(comic_id) DO UPDATE SET
  sort_key = excluded.sort_key,
  source = excluded.source,
  remote_page = excluded.remote_page,
  remote_pos = excluded.remote_pos,
  order_updated_at = datetime('now')
`, item.ID, sortKey, source.Name, page.CurrentPage, index+1).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
INSERT INTO comic_cache_state (comic_id, meta_level, cover_ready, images_total, images_local, updated_at)
VALUES (?, 1, 0, 0, 0, datetime('now'))
ON CONFLICT(comic_id) DO UPDATE SET
  meta_level = CASE WHEN comic_cache_state.meta_level < 1 THEN 1 ELSE comic_cache_state.meta_level END,
  updated_at = datetime('now')
`, item.ID).Error; err != nil {
				return err
			}
			updated++
		}
		return nil
	})
	return updated, err
}

func (store *SyncStore) UpsertComicDetail(bundle models.RemoteComicDetailBundle) error {
	return store.db.Transaction(func(tx *gorm.DB) error {
		if bundle.CategoryID != nil && *bundle.CategoryID > 0 {
			categoryName := bundle.CategoryName
			if categoryName == "" {
				categoryName = fmt.Sprintf("分类 %d", *bundle.CategoryID)
			}
			if err := tx.Exec(`
INSERT INTO categories (id, name, display_order)
VALUES (?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  name = excluded.name,
  display_order = excluded.display_order
`, *bundle.CategoryID, categoryName, int(*bundle.CategoryID)).Error; err != nil {
				return err
			}
		}

		if err := tx.Exec(`
INSERT INTO comics (
  id, title, subtitle, cover_url, rating, rating_count, favorites_remote,
  category_id, category_name, source_created_at, source_updated_at,
  source_last_seen_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'), datetime('now'))
ON CONFLICT(id) DO UPDATE SET
  title = CASE WHEN excluded.title != '' THEN excluded.title ELSE comics.title END,
  subtitle = excluded.subtitle,
  cover_url = CASE WHEN excluded.cover_url != '' THEN excluded.cover_url ELSE comics.cover_url END,
  rating = excluded.rating,
  rating_count = excluded.rating_count,
  favorites_remote = CASE
    WHEN excluded.favorites_remote > 0 OR comics.favorites_remote = 0 THEN excluded.favorites_remote
    ELSE comics.favorites_remote
  END,
  category_id = COALESCE(excluded.category_id, comics.category_id),
  category_name = CASE WHEN excluded.category_name != '' THEN excluded.category_name ELSE comics.category_name END,
  source_created_at = CASE WHEN excluded.source_created_at != '' THEN excluded.source_created_at ELSE comics.source_created_at END,
  source_updated_at = CASE WHEN excluded.source_updated_at != '' THEN excluded.source_updated_at ELSE comics.source_updated_at END,
  source_last_seen_at = datetime('now'),
  updated_at = datetime('now')
`,
			bundle.ID,
			bundle.Title,
			bundle.Subtitle,
			bundle.CoverURL,
			bundle.Rating,
			bundle.RatingCount,
			bundle.Favorites,
			bundle.CategoryID,
			bundle.CategoryName,
			bundle.SourceCreatedAt,
			bundle.SourceUpdatedAt,
		).Error; err != nil {
			return err
		}

		authorIDs := make([]int64, 0, len(bundle.Authors))
		for _, author := range bundle.Authors {
			authorID, err := store.ensureAuthor(tx, author)
			if err != nil {
				return err
			}
			authorIDs = append(authorIDs, authorID)
		}
		if err := tx.Exec(`DELETE FROM comic_authors WHERE comic_id = ?`, bundle.ID).Error; err != nil {
			return err
		}
		for index, author := range bundle.Authors {
			if err := tx.Exec(`
INSERT INTO comic_authors (comic_id, author_id, position, source_name)
VALUES (?, ?, ?, ?)
`, bundle.ID, authorIDs[index], author.Position, author.Name).Error; err != nil {
				return err
			}
		}

		if err := tx.Exec(`DELETE FROM comic_tags WHERE comic_id = ?`, bundle.ID).Error; err != nil {
			return err
		}
		for _, tag := range bundle.Tags {
			if err := tx.Exec(`
INSERT INTO tags (id, name)
VALUES (?, ?)
ON CONFLICT(id) DO UPDATE SET name = excluded.name
`, tag.ID, tag.Name).Error; err != nil {
				return err
			}
			if err := tx.Exec(`INSERT INTO comic_tags (comic_id, tag_id) VALUES (?, ?)`, bundle.ID, tag.ID).Error; err != nil {
				return err
			}
		}

		sorts := make([]int, 0, len(bundle.Images))
		for _, image := range bundle.Images {
			sorts = append(sorts, image.Sort)
			if err := tx.Exec(`
INSERT INTO comic_images (comic_id, sort, image_url, extension, local_rel_path, file_size, updated_at)
VALUES (?, ?, ?, ?, '', 0, datetime('now'))
ON CONFLICT(comic_id, sort) DO UPDATE SET
  image_url = excluded.image_url,
  extension = excluded.extension,
  updated_at = datetime('now')
`, bundle.ID, image.Sort, image.ImageURL, image.Extension).Error; err != nil {
				return err
			}
		}
		if len(sorts) == 0 {
			if err := tx.Exec(`DELETE FROM comic_images WHERE comic_id = ?`, bundle.ID).Error; err != nil {
				return err
			}
		} else {
			placeholders := strings.Repeat("?,", len(sorts))
			placeholders = strings.TrimSuffix(placeholders, ",")
			args := make([]any, 0, len(sorts)+1)
			args = append(args, bundle.ID)
			for _, sort := range sorts {
				args = append(args, sort)
			}
			if err := tx.Exec(`DELETE FROM comic_images WHERE comic_id = ? AND sort NOT IN (`+placeholders+`)`, args...).Error; err != nil {
				return err
			}
		}

		if err := tx.Exec(`
INSERT INTO comic_cache_state (
  comic_id, meta_level, cover_ready, images_total, images_local, updated_at
) VALUES (?, 2, 0, ?, 0, datetime('now'))
ON CONFLICT(comic_id) DO UPDATE SET
  meta_level = CASE WHEN comic_cache_state.meta_level < 2 THEN 2 ELSE comic_cache_state.meta_level END,
  images_total = excluded.images_total,
  updated_at = datetime('now')
`, bundle.ID, len(bundle.Images)).Error; err != nil {
			return err
		}

		return nil
	})
}

func (store *SyncStore) ensureAuthor(tx *gorm.DB, author models.RemoteAuthorRef) (int64, error) {
	normalized := normalizeName(author.Name)
	var row struct {
		ID int64 `gorm:"column:id"`
	}
	if author.ExternalID != nil && *author.ExternalID > 0 {
		if err := tx.Raw(`SELECT id FROM authors WHERE external_id = ? LIMIT 1`, *author.ExternalID).Scan(&row).Error; err != nil {
			return 0, err
		}
	}
	if row.ID == 0 && normalized != "" {
		if err := tx.Raw(`SELECT id FROM authors WHERE normalized_name = ? LIMIT 1`, normalized).Scan(&row).Error; err != nil {
			return 0, err
		}
	}
	if row.ID != 0 {
		if err := tx.Exec(`
UPDATE authors
SET external_id = COALESCE(external_id, ?),
    name = CASE WHEN ? != '' THEN ? ELSE name END,
    normalized_name = CASE WHEN ? != '' THEN ? ELSE normalized_name END
WHERE id = ?
`, author.ExternalID, author.Name, author.Name, normalized, normalized, row.ID).Error; err != nil {
			return 0, err
		}
		return row.ID, nil
	}

	result := struct {
		ID int64 `gorm:"column:id"`
	}{}
	if err := tx.Raw(`
INSERT INTO authors (external_id, name, normalized_name, created_at)
VALUES (?, ?, ?, datetime('now'))
RETURNING id
`, author.ExternalID, author.Name, normalized).Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.ID, nil
}

var nonWordPattern = regexp.MustCompile(`[^\p{L}\p{N}]+`)

func normalizeName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = nonWordPattern.ReplaceAllString(value, "")
	return value
}
