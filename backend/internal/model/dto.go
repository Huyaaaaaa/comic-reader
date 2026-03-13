package model

// ComicListItem 列表页漫画项
type ComicListItem struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	CoverURL    string  `json:"cover_url"`
	CoverBase64 string  `json:"cover_base64,omitempty"`
	Rating    float64 `json:"rating"`
	RatingCount int     `json:"rating_count"`
	Favorites   int     `json:"favorites"`
	Author      string  `json:"author"`
	AuthorID    int     `json:"author_id"`
	IsCached    bool    `json:"is_cached"`
	LocalSaved  int     `json:"local_saved"`
	LocalTotal  int     `json:"local_total"`
}

// ComicDetail 详情页漫画信息
type ComicDetail struct {
	ID           int              `json:"id"`
	Title        string           `json:"title"`
	Subtitle     string           `json:"subtitle"`
	Author       string           `json:"author"`
	AuthorID     int          `json:"author_id"`
	Authors      []AuthorInfo     `json:"authors"`
	CoverURL     string           `json:"cover_url"`
	Rating       float64          `json:"rating"`
	RatingCount  int         `json:"rating_count"`
	Favorites    int              `json:"favorites"`
	CategoryID   int              `json:"category_id"`
	CategoryName string           `json:"category_name"`
	Tags         []TagInfo        `json:"tags"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
	ReaderURL    string           `json:"reader_url"`
	IsFavorited  bool           `json:"is_favorited"`
}

// AuthorInfo 作者信息
type AuthorInfo struct {
	AuthorID   int    `json:"author_id"`
	AuthorName string `json:"author_name"`
}

// TagInfo 标签信息
type TagInfo struct {
	TagID   int    `json:"tag_id"`
	TagName string `json:"tag_name"`
}

// ImageInfo 图片信息
type ImageInfo struct {
	Sort      int    `json:"sort"`
	ComicID   int    `json:"comic_id"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
	LocalPath string `json:"local_path,omitempty"`
}

// ListResponse 列表响应
type ListResponse struct {
	Items       []ComicListItem `json:"items"`
	CurrentPage int             `json:"current_page"`
	PageSize    int             `json:"page_size"`
	TotalPages  int             `json:"total_pages"`
	Total       int64           `json:"total"`
	FromCache   bool            `json:"from_cache"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword  string `json:"keyword" form:"keyword"`
	Page     int    `json:"page" form:"page"`
	Mode     string `json:"mode" form:"mode"` // online / local
}

// FilterRequest 筛选请求
type FilterRequest struct {
	TagID      *int   `json:"tag_id" form:"tag_id"`
	CategoryID *int   `json:"category_id" form:"category_id"`
	AuthorID   *int   `json:"author_id" form:"author_id"`
	Author     string `json:"author" form:"author"`
	Page    int    `json:"page" form:"page"`
}

// SettingUpdateRequest 设置更新请求
type SettingUpdateRequest struct {
	Settings map[string]string `json:"settings"`
}

// ProgressUpdateRequest 进度更新请求
type ProgressUpdateRequest struct {
	LastPage   int     `json:"last_page"`
	TotalPages int     `json:"total_pages"`
	Progress   float64 `json:"progress"`
	ChapterID  int     `json:"chapter_id"`
}

// CacheOverview 缓存概览
type CacheOverview struct {
	TotalComics int `json:"total_comics"`
	L1Cached    int `json:"l1_cached"`
	L2Cached    int `json:"l2_cached"`
	L3Cached    int `json:"l3_cached"`
}

// HistoryWithProgress 带阅读进度的历史记录
type HistoryWithProgress struct {
	ComicID    int     `json:"comic_id"`
	Title      string  `json:"title"`
	CoverURL   string  `json:"cover_url"`
	LastReadAt string  `json:"last_read_at"`
	PageNumber int     `json:"page_number"`
	Progress   float64 `json:"progress"`
	LastPage   int     `json:"last_page"`
	TotalPages int     `json:"total_pages"`
}

// CreateDownloadRequest 创建下载任务请求
type CreateDownloadRequest struct {
	ComicID  int    `json:"comic_id" binding:"required"`
	TaskType string `json:"task_type"`
}

// DownloadPrecheckRequest 下载预检查请求
type DownloadPrecheckRequest struct {
	ComicIDs []int `json:"comic_ids" binding:"required"`
}

// DownloadPrecheckResponse 下载预检查响应
type DownloadPrecheckResponse struct {
	TargetPath     string   `json:"target_path"`
	AvailableSpace int64    `json:"available_space"`
	ReservedSpace  int64    `json:"reserved_space"`
	EstimatedSize  int64    `json:"estimated_size"`
	CanDownload    bool     `json:"can_download"`
	Warnings       []string `json:"warnings"`
}

// DownloadTaskResponse 下载任务响应（附带漫画信息）
type DownloadTaskResponse struct {
	DownloadTask
	ComicTitle    string `json:"comic_title,omitempty"`
	ComicCoverURL string `json:"comic_cover_url,omitempty"`
}

// DashboardStats 仪表盘统计
type DashboardStats struct {
	TotalComics       int `json:"total_comics"`
	CachedComics      int `json:"cached_comics"`
	CoverCached       int `json:"cover_cached"`
	TotalTags         int `json:"total_tags"`
	FavoritesCount    int `json:"favorites_count"`
	HistoryCount      int `json:"history_count"`
	DownloadingCount  int `json:"downloading_count"`
	PendingDownloads  int `json:"pending_downloads"`
}
