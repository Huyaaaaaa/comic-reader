package store

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"gorm.io/gorm"
)

type LibraryStore struct {
	db *gorm.DB
}

func NewLibraryStore(db *gorm.DB) *LibraryStore {
	return &LibraryStore{db: db}
}

func (store *LibraryStore) SetFavorite(userID int64, comicID int64, ensureOffline bool) error {
	return store.db.Exec(`
INSERT INTO user_favorites (user_id, comic_id, ensure_offline, created_at, updated_at)
VALUES (?, ?, ?, datetime('now'), datetime('now'))
ON CONFLICT(user_id, comic_id) DO UPDATE SET
  ensure_offline = excluded.ensure_offline,
  updated_at = datetime('now')
`, userID, comicID, boolToInt(ensureOffline)).Error
}

func (store *LibraryStore) RemoveFavorite(userID int64, comicID int64) error {
	return store.db.Exec(`DELETE FROM user_favorites WHERE user_id = ? AND comic_id = ?`, userID, comicID).Error
}

func (store *LibraryStore) ListFavorites(userID int64, comicID *int64, page int, pageSize int) ([]models.FavoriteItem, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	clauses := []string{"uf.user_id = ?"}
	args := []interface{}{userID}
	if comicID != nil {
		clauses = append(clauses, "uf.comic_id = ?")
		args = append(args, *comicID)
	}
	whereSQL := strings.Join(clauses, " AND ")

	var total int64
	if err := store.db.Raw(fmt.Sprintf(`SELECT COUNT(*) FROM user_favorites uf WHERE %s`, whereSQL), args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		ComicID       int64
		EnsureOffline int
		CreatedAt     string
		UpdatedAt     string
		ID            int64
		Title         string
		Subtitle      string
		CoverURL      string
		CoverLocal    string
		Rating        float64
		RatingCount   int
		Favorites     int
		CategoryID    *int64
		CategoryName  string
		MetaLevel     int
		CoverReady    int
		ImagesTotal   int
		ImagesLocal   int
	}
	queryArgs := append(append([]interface{}{}, args...), pageSize, (page-1)*pageSize)
	var rows []row
	err := store.db.Raw(fmt.Sprintf(`
SELECT
  uf.comic_id,
  uf.ensure_offline,
  uf.created_at,
  uf.updated_at,
  c.id,
  c.title,
  c.subtitle,
  c.cover_url,
  c.cover_local_rel_path AS cover_local,
  c.rating,
  c.rating_count,
  c.favorites_remote AS favorites,
  c.category_id,
  c.category_name,
  COALESCE(ccs.meta_level, 0) AS meta_level,
  COALESCE(ccs.cover_ready, 0) AS cover_ready,
  COALESCE(ccs.images_total, 0) AS images_total,
  COALESCE(ccs.images_local, 0) AS images_local
FROM user_favorites uf
JOIN comics c ON c.id = uf.comic_id
LEFT JOIN comic_cache_state ccs ON ccs.comic_id = c.id
WHERE %s
ORDER BY uf.updated_at DESC
LIMIT ? OFFSET ?
`, whereSQL), queryArgs...).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]models.FavoriteItem, 0, len(rows))
	for _, current := range rows {
		cacheState := buildCacheState(current.MetaLevel, current.CoverReady, current.ImagesTotal, current.ImagesLocal)
		items = append(items, models.FavoriteItem{
			ComicID:       current.ComicID,
			EnsureOffline: current.EnsureOffline == 1,
			OfflineReady:  cacheState.OfflineReady,
			CreatedAt:     parseSQLiteTime(current.CreatedAt),
			UpdatedAt:     parseSQLiteTime(current.UpdatedAt),
			Comic: models.ComicListItem{
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
				CacheState:   cacheState,
			},
		})
	}
	return items, int(total), nil
}

func (store *LibraryStore) UpsertHistory(userID int64, comicID int64, locator models.ReadingLocator) error {
	encoded, err := json.Marshal(locator)
	if err != nil {
		return err
	}
	return store.db.Exec(`
INSERT INTO reading_history (user_id, comic_id, locator_json, last_read_at)
VALUES (?, ?, ?, datetime('now'))
ON CONFLICT(user_id, comic_id) DO UPDATE SET
  locator_json = excluded.locator_json,
  last_read_at = datetime('now')
`, userID, comicID, string(encoded)).Error
}

func (store *LibraryStore) ListHistory(userID int64, page int, pageSize int) ([]models.HistoryItem, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	var total int64
	if err := store.db.Raw(`SELECT COUNT(*) FROM reading_history WHERE user_id = ?`, userID).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		ComicID      int64
		LocatorJSON  string
		LastReadAt   string
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
	var rows []row
	err := store.db.Raw(`
SELECT
  rh.comic_id,
  rh.locator_json,
  rh.last_read_at,
  c.id,
  c.title,
  c.subtitle,
  c.cover_url,
  c.cover_local_rel_path AS cover_local,
  c.rating,
  c.rating_count,
  c.favorites_remote AS favorites,
  c.category_id,
  c.category_name,
  COALESCE(ccs.meta_level, 0) AS meta_level,
  COALESCE(ccs.cover_ready, 0) AS cover_ready,
  COALESCE(ccs.images_total, 0) AS images_total,
  COALESCE(ccs.images_local, 0) AS images_local
FROM reading_history rh
JOIN comics c ON c.id = rh.comic_id
LEFT JOIN comic_cache_state ccs ON ccs.comic_id = c.id
WHERE rh.user_id = ?
ORDER BY rh.last_read_at DESC
LIMIT ? OFFSET ?
`, userID, pageSize, (page-1)*pageSize).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]models.HistoryItem, 0, len(rows))
	for _, current := range rows {
		cacheState := buildCacheState(current.MetaLevel, current.CoverReady, current.ImagesTotal, current.ImagesLocal)
		locator := models.ReadingLocator{}
		_ = json.Unmarshal([]byte(current.LocatorJSON), &locator)
		items = append(items, models.HistoryItem{
			ComicID:    current.ComicID,
			Locator:    locator,
			LastReadAt: parseSQLiteTime(current.LastReadAt),
			Comic: models.ComicListItem{
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
				CacheState:   cacheState,
			},
		})
	}

	return items, int(total), nil
}

func (store *LibraryStore) AddSearchHistory(userID int64, keyword string) error {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil
	}
	return store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM search_history WHERE user_id = ? AND keyword = ?`, userID, keyword).Error; err != nil {
			return err
		}
		if err := tx.Exec(`INSERT INTO search_history (user_id, keyword, searched_at) VALUES (?, ?, datetime('now'))`, userID, keyword).Error; err != nil {
			return err
		}
		return tx.Exec(`
DELETE FROM search_history
WHERE id IN (
  SELECT id FROM search_history
  WHERE user_id = ?
  ORDER BY searched_at DESC
  LIMIT -1 OFFSET 10
)
`, userID).Error
	})
}

func (store *LibraryStore) ListSearchHistory(userID int64) ([]models.SearchHistoryItem, error) {
	var rows []struct {
		Keyword    string
		SearchedAt string
	}
	if err := store.db.Raw(`
SELECT keyword, searched_at
FROM search_history
WHERE user_id = ?
ORDER BY searched_at DESC
LIMIT 10
`, userID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]models.SearchHistoryItem, 0, len(rows))
	for _, current := range rows {
		items = append(items, models.SearchHistoryItem{Keyword: current.Keyword, SearchedAt: parseSQLiteTime(current.SearchedAt)})
	}
	return items, nil
}

func (store *LibraryStore) ClearSearchHistory(userID int64) error {
	return store.db.Exec(`DELETE FROM search_history WHERE user_id = ?`, userID).Error
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
