package models

import "time"

type FavoriteItem struct {
	ComicID       int64         `json:"comic_id"`
	Comic         ComicListItem `json:"comic"`
	EnsureOffline bool          `json:"ensure_offline"`
	OfflineReady  bool          `json:"offline_ready"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type ReadingLocator struct {
	Mode        string  `json:"mode"`
	Sort        int     `json:"sort"`
	OffsetRatio float64 `json:"offset_ratio"`
}

type HistoryItem struct {
	ComicID    int64          `json:"comic_id"`
	Comic      ComicListItem  `json:"comic"`
	Locator    ReadingLocator `json:"locator"`
	LastReadAt time.Time      `json:"last_read_at"`
}

type SearchHistoryItem struct {
	Keyword    string    `json:"keyword"`
	SearchedAt time.Time `json:"searched_at"`
}
