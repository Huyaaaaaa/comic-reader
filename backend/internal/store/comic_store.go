package store

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"gorm.io/gorm"
)

type ComicStore struct {
	db *gorm.DB
}

func NewComicStore(db *gorm.DB) *ComicStore {
	return &ComicStore{db: db}
}

func (store *ComicStore) List(query models.ComicQuery) (*models.ComicListResult, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > 100 {
		pageSize = 100
	}

	clauses := []string{"1=1"}
	args := []interface{}{}

	if query.CategoryID != nil {
		clauses = append(clauses, "c.category_id = ?")
		args = append(args, *query.CategoryID)
	}
	if query.AuthorID != nil {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM comic_authors ca WHERE ca.comic_id = c.id AND ca.author_id = ?)")
		args = append(args, *query.AuthorID)
	}
	if query.TagID != nil {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM comic_tags ct WHERE ct.comic_id = c.id AND ct.tag_id = ?)")
		args = append(args, *query.TagID)
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		like := "%" + search + "%"
		clauses = append(clauses, `(c.title LIKE ? OR c.subtitle LIKE ? OR EXISTS (
      SELECT 1 FROM comic_authors ca
      JOIN authors a ON a.id = ca.author_id
      WHERE ca.comic_id = c.id AND a.name LIKE ?
    ) OR EXISTS (
      SELECT 1 FROM comic_tags ct
      JOIN tags t ON t.id = ct.tag_id
      WHERE ct.comic_id = c.id AND t.name LIKE ?
    ))`)
		args = append(args, like, like, like, like)
	}
	whereSQL := strings.Join(clauses, " AND ")

	var total int64
	if err := store.db.Raw(fmt.Sprintf(`SELECT COUNT(*) FROM comics c WHERE %s`, whereSQL), args...).Scan(&total).Error; err != nil {
		return nil, err
	}

	if query.Search == "" && query.TagID == nil && query.CategoryID == nil && query.AuthorID == nil {
		var snapshotTotal int64
		row := store.db.Raw(`SELECT total_comics FROM catalog_snapshots ORDER BY captured_at DESC LIMIT 1`).Row()
		if err := row.Scan(&snapshotTotal); err == nil && snapshotTotal > 0 {
			total = snapshotTotal
		}
	}

	offset := (page - 1) * pageSize
	listSQL := fmt.Sprintf(`
SELECT
  c.id,
  c.title,
  c.subtitle,
  c.cover_url,
  c.cover_local_rel_path,
  c.rating,
  c.rating_count,
  c.favorites_remote AS favorites,
  c.category_id,
  c.category_name,
  COALESCE(ccs.meta_level, 0) AS meta_level,
  COALESCE(ccs.cover_ready, 0) AS cover_ready,
  COALESCE(ccs.images_total, 0) AS images_total,
  COALESCE(ccs.images_local, 0) AS images_local
FROM comics c
LEFT JOIN comic_order_index coi ON coi.comic_id = c.id
LEFT JOIN comic_cache_state ccs ON ccs.comic_id = c.id
WHERE %s
ORDER BY COALESCE(coi.sort_key, 999999999) ASC, c.id DESC
LIMIT ? OFFSET ?
`, whereSQL)

	type listRow struct {
		ID           int64
		Title        string
		Subtitle     string
		CoverURL     string
		CoverLocal   string
		Rating       float64
		RatingCount  int
		Favorites    int
		CategoryID   *int64
		CategoryName string
		MetaLevel    int
		CoverReady   int
		ImagesTotal  int
		ImagesLocal  int
	}
	var rows []listRow
	queryArgs := append(append([]interface{}{}, args...), pageSize, offset)
	if err := store.db.Raw(listSQL, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]models.ComicListItem, 0, len(rows))
	for _, current := range rows {
		items = append(items, models.ComicListItem{
			ID:           current.ID,
			Title:        current.Title,
			Subtitle:     current.Subtitle,
			CoverURL:     current.CoverURL,
			CoverLocal:   current.CoverLocal,
			Rating:       current.Rating,
			RatingCount:  current.RatingCount,
			Favorites:    current.Favorites,
			CategoryID:   current.CategoryID,
			CategoryName: current.CategoryName,
			CacheState:   buildCacheState(current.MetaLevel, current.CoverReady, current.ImagesTotal, current.ImagesLocal),
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	return &models.ComicListResult{Comics: items, Total: int(total), Page: page, PageSize: pageSize, TotalPages: totalPages}, nil
}

func (store *ComicStore) Detail(id int64) (*models.ComicDetail, error) {
	type detailRow struct {
		ID           int64
		Title        string
		Subtitle     string
		CoverURL     string
		CoverLocal   string
		Rating       float64
		RatingCount  int
		Favorites    int
		CategoryID   *int64
		CategoryName string
		MetaLevel    int
		CoverReady   int
		ImagesTotal  int
		ImagesLocal  int
		CreatedAt    string
		UpdatedAt    string
	}
	var current detailRow
	err := store.db.Raw(`
SELECT
  c.id,
  c.title,
  c.subtitle,
  c.cover_url,
  c.cover_local_rel_path,
  c.rating,
  c.rating_count,
  c.favorites_remote AS favorites,
  c.category_id,
  c.category_name,
  COALESCE(ccs.meta_level, 0) AS meta_level,
  COALESCE(ccs.cover_ready, 0) AS cover_ready,
  COALESCE(ccs.images_total, 0) AS images_total,
  COALESCE(ccs.images_local, 0) AS images_local,
  c.created_at,
  c.updated_at
FROM comics c
LEFT JOIN comic_cache_state ccs ON ccs.comic_id = c.id
WHERE c.id = ?
LIMIT 1
`, id).Scan(&current).Error
	if err != nil {
		return nil, err
	}
	if current.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var authors []models.Author
	if err := store.db.Raw(`
SELECT a.id, a.external_id, a.name, a.normalized_name, ca.position
FROM comic_authors ca
JOIN authors a ON a.id = ca.author_id
WHERE ca.comic_id = ?
ORDER BY ca.position ASC, a.id ASC
`, id).Scan(&authors).Error; err != nil {
		return nil, err
	}

	var tags []models.Tag
	if err := store.db.Raw(`
SELECT t.id, t.name
FROM comic_tags ct
JOIN tags t ON t.id = ct.tag_id
WHERE ct.comic_id = ?
ORDER BY t.name ASC
`, id).Scan(&tags).Error; err != nil {
		return nil, err
	}

	createdAt := parseSQLiteTime(current.CreatedAt)
	updatedAt := parseSQLiteTime(current.UpdatedAt)

	return &models.ComicDetail{
		ID:           current.ID,
		Title:        current.Title,
		Subtitle:     current.Subtitle,
		CoverURL:     current.CoverURL,
		CoverLocal:   current.CoverLocal,
		Rating:       current.Rating,
		RatingCount:  current.RatingCount,
		Favorites:    current.Favorites,
		CategoryID:   current.CategoryID,
		CategoryName: current.CategoryName,
		Authors:      authors,
		Tags:         tags,
		ImagesTotal:  current.ImagesTotal,
		CacheState:   buildCacheState(current.MetaLevel, current.CoverReady, current.ImagesTotal, current.ImagesLocal),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

func (store *ComicStore) Images(id int64) (*models.ComicImagesResult, error) {
	type imageRow struct {
		ComicID      int64
		Sort         int
		ImageURL     string
		Extension    string
		LocalRelPath string
		FileSize     int64
		CachedInt    int `gorm:"column:cached"`
	}
	var rows []imageRow
	if err := store.db.Raw(`
SELECT comic_id, sort, image_url, extension, local_rel_path, file_size,
  CASE WHEN TRIM(COALESCE(local_rel_path, '')) != '' THEN 1 ELSE 0 END AS cached
FROM comic_images
WHERE comic_id = ?
ORDER BY sort ASC
`, id).Scan(&rows).Error; err != nil {
		return nil, err
	}
	images := make([]models.ComicImage, 0, len(rows))
	for _, row := range rows {
		images = append(images, models.ComicImage{
			ComicID:      row.ComicID,
			Sort:         row.Sort,
			ImageURL:     row.ImageURL,
			Extension:    row.Extension,
			LocalRelPath: row.LocalRelPath,
			FileSize:     row.FileSize,
			Cached:       row.CachedInt == 1,
		})
	}
	return &models.ComicImagesResult{Images: images, Total: len(images)}, nil
}

func (store *ComicStore) GetImageTarget(comicID int64, sort int) (*models.ComicImageTarget, error) {
	var row models.ComicImageTarget
	err := store.db.Raw(`
SELECT comic_id, sort, image_url, extension, local_rel_path
FROM comic_images
WHERE comic_id = ? AND sort = ?
LIMIT 1
`, comicID, sort).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ComicID == 0 && strings.TrimSpace(row.ImageURL) == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (store *ComicStore) MarkImageCached(comicID int64, sort int, localRelPath string, fileSize int64) error {
	return store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
UPDATE comic_images
SET local_rel_path = ?,
    file_size = ?,
    cached_at = datetime('now'),
    updated_at = datetime('now')
WHERE comic_id = ? AND sort = ?
`, localRelPath, fileSize, comicID, sort).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
INSERT INTO comic_cache_state (comic_id, meta_level, cover_ready, images_total, images_local, first_collected_at, updated_at)
VALUES (
  ?,
  2,
  0,
  (SELECT COUNT(*) FROM comic_images WHERE comic_id = ?),
  (SELECT COUNT(*) FROM comic_images WHERE comic_id = ? AND TRIM(COALESCE(local_rel_path, '')) != ''),
  datetime('now'),
  datetime('now')
)
ON CONFLICT(comic_id) DO UPDATE SET
  meta_level = CASE WHEN comic_cache_state.meta_level < 2 THEN 2 ELSE comic_cache_state.meta_level END,
  images_total = (SELECT COUNT(*) FROM comic_images WHERE comic_id = excluded.comic_id),
  images_local = (SELECT COUNT(*) FROM comic_images WHERE comic_id = excluded.comic_id AND TRIM(COALESCE(local_rel_path, '')) != ''),
  first_collected_at = COALESCE(comic_cache_state.first_collected_at, datetime('now')),
  fully_cached_at = CASE
    WHEN (SELECT COUNT(*) FROM comic_images WHERE comic_id = excluded.comic_id) > 0
     AND (SELECT COUNT(*) FROM comic_images WHERE comic_id = excluded.comic_id AND TRIM(COALESCE(local_rel_path, '')) != '') = (SELECT COUNT(*) FROM comic_images WHERE comic_id = excluded.comic_id)
    THEN datetime('now')
    ELSE comic_cache_state.fully_cached_at
  END,
  updated_at = datetime('now')
`, comicID, comicID, comicID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (store *ComicStore) Tags() ([]models.Tag, error) {
	var tags []models.Tag
	if err := store.db.Raw(`SELECT id, name FROM tags ORDER BY name ASC`).Scan(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (store *ComicStore) Categories() ([]models.Category, error) {
	var categories []models.Category
	if err := store.db.Raw(`SELECT id, name, display_order FROM categories ORDER BY display_order ASC, id ASC`).Scan(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (store *ComicStore) Authors(page int, pageSize int, search string) ([]models.Author, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	clauses := []string{"1=1"}
	args := []interface{}{}
	if strings.TrimSpace(search) != "" {
		like := "%" + strings.TrimSpace(search) + "%"
		clauses = append(clauses, "name LIKE ?")
		args = append(args, like)
	}
	whereSQL := strings.Join(clauses, " AND ")
	var total int64
	if err := store.db.Raw(fmt.Sprintf(`SELECT COUNT(*) FROM authors WHERE %s`, whereSQL), args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]interface{}{}, args...), pageSize, (page-1)*pageSize)
	var authors []models.Author
	if err := store.db.Raw(fmt.Sprintf(`SELECT id, external_id, name, normalized_name FROM authors WHERE %s ORDER BY name ASC LIMIT ? OFFSET ?`, whereSQL), queryArgs...).Scan(&authors).Error; err != nil {
		return nil, 0, err
	}
	return authors, int(total), nil
}

func buildCacheState(metaLevel int, coverReady int, imagesTotal int, imagesLocal int) models.CacheStateSummary {
	return models.CacheStateSummary{
		MetaLevel:    metaLevel,
		CoverReady:   coverReady == 1,
		ImagesTotal:  imagesTotal,
		ImagesLocal:  imagesLocal,
		OfflineReady: metaLevel >= 2 && coverReady == 1 && imagesTotal > 0 && imagesTotal == imagesLocal,
	}
}

func parseSQLiteTime(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", time.RFC3339Nano}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
