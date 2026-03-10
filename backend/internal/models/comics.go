package models

import "time"

type Category struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
}

type Author struct {
	ID             int64  `json:"id"`
	ExternalID     *int64 `json:"external_id,omitempty"`
	Name           string `json:"name"`
	NormalizedName string `json:"normalized_name,omitempty"`
	Position       int    `json:"position,omitempty"`
}

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CacheStateSummary struct {
	MetaLevel    int  `json:"meta_level"`
	CoverReady   bool `json:"cover_ready"`
	ImagesTotal  int  `json:"images_total"`
	ImagesLocal  int  `json:"images_local"`
	OfflineReady bool `json:"offline_ready"`
}

type ComicListItem struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Subtitle     string            `json:"subtitle"`
	CoverURL     string            `json:"cover_url"`
	CoverLocal   string            `json:"cover_local_rel_path"`
	Rating       float64           `json:"rating"`
	RatingCount  int               `json:"rating_count"`
	Favorites    int               `json:"favorites"`
	CategoryID   *int64            `json:"category_id"`
	CategoryName string            `json:"category_name"`
	CacheState   CacheStateSummary `json:"cache_state"`
}

type ComicListResult struct {
	Comics     []ComicListItem `json:"comics"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

type ComicDetail struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Subtitle     string            `json:"subtitle"`
	CoverURL     string            `json:"cover_url"`
	CoverLocal   string            `json:"cover_local_rel_path"`
	Rating       float64           `json:"rating"`
	RatingCount  int               `json:"rating_count"`
	Favorites    int               `json:"favorites"`
	CategoryID   *int64            `json:"category_id"`
	CategoryName string            `json:"category_name"`
	Authors      []Author          `json:"authors"`
	Tags         []Tag             `json:"tags"`
	ImagesTotal  int               `json:"images_total"`
	CacheState   CacheStateSummary `json:"cache_state"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type ComicImage struct {
	ComicID      int64  `json:"comic_id"`
	Sort         int    `json:"sort"`
	ImageURL     string `json:"image_url"`
	Extension    string `json:"extension"`
	LocalRelPath string `json:"local_rel_path"`
	FileSize     int64  `json:"file_size"`
	Cached       bool   `json:"cached"`
}

type ComicImagesResult struct {
	Images []ComicImage `json:"images"`
	Total  int          `json:"total"`
}

type ComicImageTarget struct {
	ComicID      int64
	Sort         int
	ImageURL     string
	Extension    string
	LocalRelPath string
}

type ComicCoverTarget struct {
	ComicID      int64
	CoverURL     string
	LocalRelPath string
}

type ComicQuery struct {
	Page       int
	PageSize   int
	Search     string
	TagID      *int64
	CategoryID *int64
	AuthorID   *int64
}
