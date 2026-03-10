package store

import (
	"fmt"
	"time"

	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"gorm.io/gorm"
)

type SettingsStore struct {
	db *gorm.DB
}

func NewSettingsStore(db *gorm.DB) *SettingsStore {
	return &SettingsStore{db: db}
}

func (store *SettingsStore) List() (map[string]models.AppSetting, error) {
	var rows []struct {
		Key          string
		ValueJSON    string
		ValueType    string
		DefaultValue string
		Description  string
		UpdatedAt    string
	}
	if err := store.db.Raw(`
SELECT key, value_json, value_type, default_value, description, updated_at
FROM app_settings
ORDER BY key ASC
`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]models.AppSetting, len(rows))
	for _, row := range rows {
		result[row.Key] = models.AppSetting{
			Key:          row.Key,
			ValueJSON:    row.ValueJSON,
			ValueType:    row.ValueType,
			DefaultValue: row.DefaultValue,
			Description:  row.Description,
			UpdatedAt:    parseSQLiteTime(row.UpdatedAt),
		}
	}
	return result, nil
}

func (store *SettingsStore) Update(key string, valueJSON string) error {
	result := store.db.Model(&models.AppSetting{}).
		Where("key = ?", key).
		Updates(map[string]any{"value_json": valueJSON, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("setting not found")
	}
	return nil
}
