package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"github.com/huyaaaaaa/hehuan-reader/internal/store"
)

type SourceHealthService struct {
	store      *store.SourceStore
	settings   *store.SettingsStore
	events     *EventBroker
	httpClient *http.Client
}

func NewSourceHealthService(sourceStore *store.SourceStore, settingsStore *store.SettingsStore, events *EventBroker) *SourceHealthService {
	return &SourceHealthService{
		store:      sourceStore,
		settings:   settingsStore,
		events:     events,
		httpClient: &http.Client{},
	}
}

func (service *SourceHealthService) CheckSource(ctx context.Context, item *models.SourceSite, checkType string) (*models.SourceHealthCheck, error) {
	settingsMap, err := service.settings.List()
	if err != nil {
		return nil, err
	}
	attemptTimeout := parseDurationSeconds(settingsMap["source_request_timeout_seconds"].ValueJSON, 15)
	failureThreshold := parseInt(settingsMap["source_failure_threshold"].ValueJSON, 3)
	check := &models.SourceHealthCheck{SourceID: item.ID, CheckType: checkType, Status: "ok", CheckedAt: time.Now()}

	for attempt := 1; attempt <= failureThreshold; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
		request, err := http.NewRequestWithContext(attemptCtx, http.MethodGet, normalizeURL(item.BaseURL), nil)
		if err != nil {
			cancel()
			return nil, err
		}

		startedAt := time.Now()
		response, requestErr := service.httpClient.Do(request)
		cancel()
		if requestErr == nil && response != nil {
			_ = response.Body.Close()
			latency := int(time.Since(startedAt).Milliseconds())
			check.LatencyMS = &latency
			check.Status = "ok"
			check.ErrorMessage = ""
			item.Status = "available"
			item.ConsecutiveFailures = 0
			item.LastLatencyMS = &latency
			now := time.Now()
			item.LastCheckedAt = &now
			item.LastError = ""
			if err := service.store.Update(item.ID, map[string]any{
				"status":               item.Status,
				"consecutive_failures": item.ConsecutiveFailures,
				"last_latency_ms":      latency,
				"last_checked_at":      now,
				"last_error":           item.LastError,
			}); err != nil {
				return nil, err
			}
			if err := service.store.AddCheck(check); err != nil {
				return nil, err
			}
			service.events.Publish("source.check.ok", map[string]any{"source_id": item.ID, "attempt": attempt, "latency_ms": latency})
			return check, nil
		}

		item.ConsecutiveFailures++
		check.Status = "failed"
		if requestErr != nil {
			check.ErrorMessage = requestErr.Error()
		} else {
			check.ErrorMessage = "unknown source error"
		}
		service.events.Publish("source.check.retry", map[string]any{"source_id": item.ID, "attempt": attempt, "error": check.ErrorMessage})
	}

	now := time.Now()
	item.Status = "unavailable"
	item.LastCheckedAt = &now
	item.LastError = check.ErrorMessage
	if err := service.store.Update(item.ID, map[string]any{
		"status":               item.Status,
		"consecutive_failures": item.ConsecutiveFailures,
		"last_checked_at":      now,
		"last_error":           item.LastError,
	}); err != nil {
		return nil, err
	}
	if err := service.store.AddCheck(check); err != nil {
		return nil, err
	}
	service.events.Publish("source.check.failed", map[string]any{"source_id": item.ID, "error": check.ErrorMessage})
	return check, errors.New(check.ErrorMessage)
}

func parseDurationSeconds(raw string, fallback int) time.Duration {
	value := parseInt(raw, fallback)
	return time.Duration(value) * time.Second
}

func parseInt(raw string, fallback int) int {
	var parsed int
	if _, err := fmt.Sscanf(strings.Trim(raw, "\""), "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func normalizeURL(rawURL string) string {
	trimmed := strings.TrimSpace(rawURL)
	if strings.Contains(trimmed, "://") {
		return trimmed
	}
	return (&neturl.URL{Scheme: "https", Host: trimmed}).String()
}
