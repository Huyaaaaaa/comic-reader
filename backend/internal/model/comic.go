package model

import (
	"time"
)

// Comic 漫画主表
type Comic struct {
	ID              int       `gorm:"primaryKey" json:"id"`
	Title           string    `gorm:"type:text" json:"title"`
	Subtitle   string    `gorm:"type:text" json:"subtitle"`
	Author          string    `gorm:"type:text" json:"author"`
	AuthorID        int       `json:"author_id"`
	CoverURL        string    `gorm:"type:text" json:"cover_url"`
	CoverBase64     string    `gorm:"type:text" json:"cover_base64,omitempty"`
	Rating          float64   `json:"rating"`
	RatingCount     int       `json:"rating_count"`
	Favorites       int    `json:"favorites"`
	CategoryID      int       `json:"category_id"`
	CategoryName    string    `json:"category_name"`
	Description     string    `gorm:"type:text" json:"description"`
	PageCount       int       `json:"page_count"`
	Status          string    `gorm:"type:text;default:unknown" json:"status"` // ongoing/completed/unknown
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	HasCoverCached  bool      `json:"has_cover_cached"`
	CoverCachedAt   *time.Time `json:"cover_cached_at,omitempty"`
	CoverSize       int       `json:"cover_size"`
}

// ComicImage 漫画图片
type ComicImage struct {
	ComicID   int    `gorm:"primaryKey" json:"comic_id"`
	Sort      int    `gorm:"primaryKey" json:"sort"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
	URL       string `gorm:"type:text" json:"url"`
	LocalPath string `gorm:"type:text" json:"local_path"`
}

// Tag 标签表
type Tag struct {
	ID   int    `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:text;not null" json:"name"`
}

// ComicTag 漫画标签关联
type ComicTag struct {
	ComicID int `gorm:"primaryKey" json:"comic_id"`
	TagID   int `gorm:"primaryKey" json:"tag_id"`
}

// ComicAuthor 漫画作者关联
type ComicAuthor struct {
	ComicID    int    `gorm:"primaryKey" json:"comic_id"`
	AuthorID   int    `gorm:"primaryKey" json:"author_id"`
	AuthorName string `gorm:"primaryKey;type:text" json:"author_name"`
	Position   int    `json:"position"`
}

// ComicMetadataState 漫画元数据状态
type ComicMetadataState struct {
	ComicID            int        `gorm:"primaryKey" json:"comic_id"`
	ListCached           bool       `json:"list_cached"`
	DetailCached       bool       `json:"detail_cached"`
	AuthorReady          bool       `json:"author_ready"`
	TagsReady            bool       `json:"tags_ready"`
	LastListCachedAt     *time.Time `json:"last_list_cached_at,omitempty"`
	LastDetailCachedAt   *time.Time `json:"last_detail_cached_at,omitempty"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// ReadingHistory 阅读历史
type ReadingHistory struct {
	ComicID    int       `gorm:"primaryKey" json:"comic_id"`
	Title    string    `gorm:"type:text" json:"title"`
	CoverURL   string    `gorm:"type:text" json:"cover_url"`
	PageNumber int       `json:"page_number"`
	Progress   float64   `json:"progress"`
	LastReadAt time.Time `json:"last_read_at"`
}

// ComicFavorite 收藏记录
type ComicFavorite struct {
	ComicID     int       `gorm:"primaryKey" json:"comic_id"`
	Title       string    `gorm:"type:text" json:"title"`
	CoverURL    string    `gorm:"type:text" json:"cover_url"`
	FavoritedAt time.Time `json:"favorited_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ComicListCache 列表缓存
type ComicListCache struct {
	ComicID     int       `gorm:"primaryKey" json:"comic_id"`
	Page        int       `gorm:"primaryKey" json:"page"`
	Position    int       `json:"position"`
	Title    string    `gorm:"type:text" json:"title"`
	Author      string    `gorm:"type:text" json:"author"`
	AuthorID    int       `json:"author_id"`
	CoverURL    string    `gorm:"type:text" json:"cover_url"`
	Rating      float64   `json:"rating"`
	RatingCount int       `json:"rating_count"`
	Favorites   int       `json:"favorites"`
	CachedAt    time.Time `json:"cached_at"`
}

// DownloadTask 下载任务（v2: 7状态 + 锁机制 + 断点续传）
type DownloadTask struct {
	ID             int        `gorm:"primaryKey;autoIncrement" json:"id"`
	ComicID        int        `gorm:"index" json:"comic_id"`
	UserID         int        `gorm:"default:1" json:"user_id"`
	Title          string     `gorm:"type:text" json:"title"`
	TaskType       string     `gorm:"type:text;default:images" json:"task_type"`
	TriggerSource  string     `gorm:"type:text;default:manual" json:"trigger_source"`
	Status         string     `gorm:"type:text;default:queued;index" json:"status"` // queued/downloading/verifying/completed/failed/paused/canceled
	Priority       int        `gorm:"default:0" json:"priority"`
	Total          int        `json:"total"`
	Current        int        `json:"current"`
	RetryCount     int        `gorm:"default:0" json:"retry_count"`
	LockToken      string     `gorm:"type:text" json:"lock_token,omitempty"`
	LockAcquiredAt *time.Time `json:"lock_acquired_at,omitempty"`
	LastError      string     `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
}

// CoverCacheQueue 封面缓存队列
type CoverCacheQueue struct {
	ComicID    int       `gorm:"primaryKey" json:"comic_id"`
	Title      string    `gorm:"type:text" json:"title"`
	CoverURL   string    `gorm:"type:text" json:"cover_url"`
	Priority   int       `json:"priority"`
	Status     string    `json:"status"` // pending / processing / completed / failed
	RetryCount int       `json:"retry_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SystemMetadata 系统元数据
type SystemMetadata struct {
	Key       string    `gorm:"primaryKey;type:text" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchHistory 搜索历史
type SearchHistory struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int       `gorm:"default:1" json:"user_id"`
	Keyword    string    `gorm:"type:text;not null" json:"keyword"`
	ResultCount int      `json:"result_count"`
	SearchedAt time.Time `json:"searched_at"`
}

// UserSetting 用户设置
type UserSetting struct {
	UserID    int       `gorm:"primaryKey" json:"user_id"`
	Key       string    `gorm:"primaryKey;type:text" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CacheState 缓存状态
type CacheState struct {
	ComicID    int       `gorm:"primaryKey" json:"comic_id"`
	L1Cached   bool      `json:"l1_cached"`
	L2Cached   bool      `json:"l2_cached"`
	L3Cached   bool      `json:"l3_cached"`
	L3Progress float64   `json:"l3_progress"`
	L3Total    int       `json:"l3_total"`
	L3Current  int       `json:"l3_current"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ReadingProgress 阅读进度
type ReadingProgress struct {
	ComicID    int       `gorm:"primaryKey" json:"comic_id"`
	UserID     int       `gorm:"primaryKey;default:1" json:"user_id"`
	LastPage   int       `json:"last_page"`
	TotalPages int       `json:"total_pages"`
	Progress   float64   `json:"progress"`
	ChapterID  int       `json:"chapter_id"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SourceSite 源站
type SourceSite struct {
	ID        int        `gorm:"primaryKey;autoIncrement" json:"id"`
	URL       string     `gorm:"type:text;not null" json:"url"`
	Name      string     `gorm:"type:text" json:"name"`
	ImageCDN  string     `gorm:"type:text" json:"image_cdn"`
	Priority  int        `json:"priority"`
	Status    string     `gorm:"type:text;default:active" json:"status"` // active/inactive/failed
	FailCount int        `gorm:"default:0" json:"fail_count"`
	Latency   int        `json:"latency"`
	LastCheck *time.Time `json:"last_check,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ImportExportJob 导入导出任务
type ImportExportJob struct {
	ID            int        `gorm:"primaryKey;autoIncrement" json:"id"`
	Direction     string     `gorm:"type:text;not null" json:"direction"`                // import/export
	Scope         string     `gorm:"type:text;not null" json:"scope"`                    // single_comic/selected_offline_ready/all_cached_comics/all_covers/all_images
	Status        string     `gorm:"type:text;not null;default:queued" json:"status"`    // queued/running/completed/failed/canceled
	FilePath      string     `gorm:"type:text" json:"file_path"`
	OptionsJSON   string     `gorm:"type:text;default:{}" json:"options_json"`
	SummaryJSON   string     `gorm:"type:text;default:{}" json:"summary_json"`
	ConflictCount int        `json:"conflict_count"`
	LastError     string     `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
}

// TableName 指定表名
func (Comic) TableName() string              { return "comics" }
func (ComicImage) TableName() string         { return "comic_images" }
func (Tag) TableName() string                { return "tags" }
func (ComicTag) TableName() string           { return "comic_tags" }
func (ComicAuthor) TableName() string        { return "comic_authors" }
func (ComicMetadataState) TableName() string { return "comic_metadata_state" }
func (ReadingHistory) TableName() string     { return "reading_history" }
func (ComicFavorite) TableName() string      { return "comic_favorites" }
func (ComicListCache) TableName() string     { return "comic_list_cache" }
func (DownloadTask) TableName() string       { return "download_tasks" }
func (CoverCacheQueue) TableName() string    { return "cover_cache_queue" }
func (SystemMetadata) TableName() string     { return "system_metadata" }
func (SearchHistory) TableName() string      { return "search_histories" }
func (UserSetting) TableName() string        { return "user_settings" }
func (CacheState) TableName() string         { return "cache_states" }
func (ReadingProgress) TableName() string    { return "reading_progress" }
func (SourceSite) TableName() string         { return "source_sites" }
// UpdateRecord 更新记录
type UpdateRecord struct {
	ID             int        `gorm:"primaryKey;autoIncrement" json:"id"`
	UpdateType     string     `gorm:"type:text;not null" json:"update_type"`   // content/app
	Channel        string     `gorm:"type:text" json:"channel"`
	RemoteVersion  string     `gorm:"type:text" json:"remote_version"`
	CurrentVersion string     `gorm:"type:text" json:"current_version"`
	HasUpdate      bool       `json:"has_update"`
	CheckMode      string     `gorm:"type:text" json:"check_mode"` // manual/auto
	Result         string     `gorm:"type:text" json:"result"`     // JSON summary
	CheckedAt      time.Time  `json:"checked_at"`
	AppliedAt      *time.Time `json:"applied_at,omitempty"`
}

func (ImportExportJob) TableName() string    { return "import_export_jobs" }
func (UpdateRecord) TableName() string       { return "update_records" }
