package config

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	AppName        string
	AppVersion     string
	HTTPAddr       string
	DataDir        string
	DBPath         string
	AllowedOrigins []string
}

func Load() Config {
	dataDir := getenv("HEHUAN_DATA_DIR", "./data")
	dbPath := getenv("HEHUAN_DB_PATH", filepath.Join(dataDir, "hehuan.db"))

	return Config{
		AppName:        getenv("HEHUAN_APP_NAME", "合欢阅读器"),
		AppVersion:     getenv("HEHUAN_APP_VERSION", "0.1.0"),
		HTTPAddr:       getenv("HEHUAN_HTTP_ADDR", ":8080"),
		DataDir:        dataDir,
		DBPath:         dbPath,
		AllowedOrigins: parseOrigins(os.Getenv("HEHUAN_ALLOWED_ORIGINS")),
	}
}

func getenv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
