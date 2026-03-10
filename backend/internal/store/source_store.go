package store

import (
	"time"

	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"gorm.io/gorm"
)

type SourceStore struct {
	db *gorm.DB
}

func NewSourceStore(db *gorm.DB) *SourceStore {
	return &SourceStore{db: db}
}

func (store *SourceStore) List() ([]models.SourceSite, error) {
	var rows []struct {
		ID                  int64
		Name                string
		BaseURL             string
		NavigatorURL        string
		Priority            int
		Enabled             int
		Status              string
		LastLatencyMS       *int
		LastCheckedAt       *string
		ConsecutiveFailures int
		LastError           string
		CreatedAt           string
		UpdatedAt           string
	}
	if err := store.db.Raw(`
SELECT id, name, base_url, navigator_url, priority, enabled, status,
       last_latency_ms, last_checked_at, consecutive_failures, last_error,
       created_at, updated_at
FROM source_sites
ORDER BY priority ASC, id ASC
`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]models.SourceSite, 0, len(rows))
	for _, row := range rows {
		var lastCheckedAt *time.Time
		if row.LastCheckedAt != nil && *row.LastCheckedAt != "" {
			parsed := parseSQLiteTime(*row.LastCheckedAt)
			lastCheckedAt = &parsed
		}
		items = append(items, models.SourceSite{
			ID:                  row.ID,
			Name:                row.Name,
			BaseURL:             row.BaseURL,
			NavigatorURL:        row.NavigatorURL,
			Priority:            row.Priority,
			Enabled:             row.Enabled == 1,
			Status:              row.Status,
			LastLatencyMS:       row.LastLatencyMS,
			LastCheckedAt:       lastCheckedAt,
			ConsecutiveFailures: row.ConsecutiveFailures,
			LastError:           row.LastError,
			CreatedAt:           parseSQLiteTime(row.CreatedAt),
			UpdatedAt:           parseSQLiteTime(row.UpdatedAt),
		})
	}
	return items, nil
}

func (store *SourceStore) Create(item *models.SourceSite) error {
	return store.db.Create(item).Error
}

func (store *SourceStore) Update(id int64, changes map[string]any) error {
	changes["updated_at"] = time.Now()
	return store.db.Model(&models.SourceSite{}).Where("id = ?", id).Updates(changes).Error
}

func (store *SourceStore) Delete(id int64) error {
	return store.db.Delete(&models.SourceSite{}, id).Error
}

func (store *SourceStore) Get(id int64) (*models.SourceSite, error) {
	items, err := store.List()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ID == id {
			copyItem := item
			return &copyItem, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (store *SourceStore) AddCheck(check *models.SourceHealthCheck) error {
	return store.db.Create(check).Error
}
