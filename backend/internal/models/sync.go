package models

import "time"

type RemoteComicListItem struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	CoverURL    string  `json:"cover_url"`
	Rating      float64 `json:"rating"`
	RatingCount int     `json:"rating_count"`
	Favorites   int     `json:"favorites"`
}

type RemoteListPage struct {
	CurrentPage int                   `json:"current_page"`
	TotalPages  int                   `json:"total_pages"`
	Items       []RemoteComicListItem `json:"items"`
}

type RemoteAuthorRef struct {
	ExternalID *int64 `json:"external_id,omitempty"`
	Name       string `json:"name"`
	Position   int    `json:"position"`
}

type RemoteTagRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type RemoteComicImage struct {
	Sort      int    `json:"sort"`
	ImageURL  string `json:"image_url"`
	Extension string `json:"extension"`
}

type RemoteComicDetailBundle struct {
	ID              int64              `json:"id"`
	Title           string             `json:"title"`
	Subtitle        string             `json:"subtitle"`
	CoverURL        string             `json:"cover_url"`
	Rating          float64            `json:"rating"`
	RatingCount     int                `json:"rating_count"`
	Favorites       int                `json:"favorites"`
	CategoryID      *int64             `json:"category_id,omitempty"`
	CategoryName    string             `json:"category_name"`
	Authors         []RemoteAuthorRef  `json:"authors"`
	Tags            []RemoteTagRef     `json:"tags"`
	Images          []RemoteComicImage `json:"images"`
	SourceCreatedAt string             `json:"source_created_at"`
	SourceUpdatedAt string             `json:"source_updated_at"`
}

type SyncHeadResult struct {
	SourceID      int64     `json:"source_id"`
	SourceName    string    `json:"source_name"`
	ScannedPages  int       `json:"scanned_pages"`
	TotalPages    int       `json:"total_pages"`
	ScannedItems  int       `json:"scanned_items"`
	UpdatedItems  int       `json:"updated_items"`
	TotalComics   int       `json:"total_comics"`
	LastPageCount int       `json:"last_page_count"`
	CapturedAt    time.Time `json:"captured_at"`
}

type SyncComicResult struct {
	SourceID     int64     `json:"source_id"`
	SourceName   string    `json:"source_name"`
	ComicID      int64     `json:"comic_id"`
	Title        string    `json:"title"`
	ImagesTotal  int       `json:"images_total"`
	AuthorsTotal int       `json:"authors_total"`
	TagsTotal    int       `json:"tags_total"`
	CapturedAt   time.Time `json:"captured_at"`
}
