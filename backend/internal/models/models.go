package models

import "time"

type AppSetting struct {
	Key          string    `gorm:"column:key;primaryKey" json:"key"`
	ValueJSON    string    `gorm:"column:value_json;not null" json:"value_json"`
	ValueType    string    `gorm:"column:value_type;not null" json:"value_type"`
	DefaultValue string    `gorm:"column:default_value;not null" json:"default_value"`
	Description  string    `gorm:"column:description;not null" json:"description"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

func (AppSetting) TableName() string { return "app_settings" }

type SourceSite struct {
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name                string     `gorm:"column:name;not null" json:"name"`
	BaseURL             string     `gorm:"column:base_url;not null" json:"base_url"`
	NavigatorURL        string     `gorm:"column:navigator_url" json:"navigator_url"`
	Priority            int        `gorm:"column:priority;not null" json:"priority"`
	Enabled             bool       `gorm:"column:enabled;not null" json:"enabled"`
	Status              string     `gorm:"column:status;not null" json:"status"`
	LastLatencyMS       *int       `gorm:"column:last_latency_ms" json:"last_latency_ms"`
	LastCheckedAt       *time.Time `gorm:"column:last_checked_at" json:"last_checked_at"`
	ConsecutiveFailures int        `gorm:"column:consecutive_failures;not null" json:"consecutive_failures"`
	LastError           string     `gorm:"column:last_error;not null" json:"last_error"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
}

func (SourceSite) TableName() string { return "source_sites" }

type SourceHealthCheck struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SourceID     int64     `gorm:"column:source_id;not null" json:"source_id"`
	CheckType    string    `gorm:"column:check_type;not null" json:"check_type"`
	Status       string    `gorm:"column:status;not null" json:"status"`
	LatencyMS    *int      `gorm:"column:latency_ms" json:"latency_ms"`
	ErrorMessage string    `gorm:"column:error_message;not null" json:"error_message"`
	CheckedAt    time.Time `gorm:"column:checked_at;not null" json:"checked_at"`
}

func (SourceHealthCheck) TableName() string { return "source_health_checks" }

type CatalogSnapshot struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SourceID      *int64    `gorm:"column:source_id" json:"source_id"`
	TotalComics   int       `gorm:"column:total_comics;not null" json:"total_comics"`
	TotalPages    int       `gorm:"column:total_pages;not null" json:"total_pages"`
	LastPageCount int       `gorm:"column:last_page_count;not null" json:"last_page_count"`
	CapturedAt    time.Time `gorm:"column:captured_at;not null" json:"captured_at"`
}

func (CatalogSnapshot) TableName() string { return "catalog_snapshots" }
