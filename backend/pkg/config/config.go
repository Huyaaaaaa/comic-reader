package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Crawler  CrawlerConfig  `mapstructure:"crawler"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Download DownloadConfig `mapstructure:"download"`
	Log    LogConfig      `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port        int      `mapstructure:"port"`
	Mode        string   `mapstructure:"mode"`
	CorsOrigins []string `mapstructure:"cors_origins"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path         string `mapstructure:"path"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// CrawlerConfig 爬虫配置
type CrawlerConfig struct {
	BaseURLs        []string `mapstructure:"base_urls"`
	ActiveBaseURL   string   `mapstructure:"active_base_url"`
	ImageCDN        string   `mapstructure:"image_cdn"`
	RequestDelayMin float64  `mapstructure:"request_delay_min"`
	RequestDelayMax float64  `mapstructure:"request_delay_max"`
	AntiBanMode     string   `mapstructure:"anti_ban_mode"`
	UserAgents      []string `mapstructure:"user_agents"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	CoverMax     int    `mapstructure:"cover_max"`
	CoverTargetSizeKB int    `mapstructure:"cover_target_size_kb"`
	CoverStrategy     string `mapstructure:"cover_strategy"`
	ListCacheEnabled  bool   `mapstructure:"list_cache_enabled"`
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	Dir         string  `mapstructure:"dir"`
	Concurrency int     `mapstructure:"concurrency"`
	DelayMin    float64 `mapstructure:"delay_min"`
	DelayMax    float64 `mapstructure:"delay_max"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level   string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.path", "./data/comics.db")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("crawler.request_delay_min", 2.0)
	viper.SetDefault("crawler.request_delay_max", 5.0)
	viper.SetDefault("crawler.anti_ban_mode", "safe")
	viper.SetDefault("cache.cover_max", 2000)
	viper.SetDefault("cache.cover_target_size_kb", 50)
	viper.SetDefault("cache.cover_strategy", "passive")
	viper.SetDefault("cache.list_cache_enabled", true)
	viper.SetDefault("download.dir", "./downloads")
	viper.SetDefault("download.concurrency", 3)
	viper.SetDefault("download.delay_min", 1.0)
	viper.SetDefault("download.delay_max", 3.0)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "./logs/app.log")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 3)
	viper.SetDefault("log.max_age", 28)
}
