package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"github.com/huyaaaaaa/hehuan-reader/internal/store"
)

type SourceClient struct {
	store      *store.SourceStore
	settings   *store.SettingsStore
	events     *EventBroker
	httpClient *http.Client
	parser     *monsterParser
}

func NewSourceClient(sourceStore *store.SourceStore, settingsStore *store.SettingsStore, events *EventBroker) *SourceClient {
	return &SourceClient{
		store:      sourceStore,
		settings:   settingsStore,
		events:     events,
		httpClient: &http.Client{},
		parser:     newMonsterParser(),
	}
}

func (client *SourceClient) FetchHeadPage(ctx context.Context, preferredSourceID *int64, page int) (*models.SourceSite, *models.RemoteListPage, error) {
	if page <= 0 {
		page = 1
	}
	var result *models.RemoteListPage
	source, err := client.withSource(ctx, preferredSourceID, fmt.Sprintf("sync.head.page.%d", page), func(runCtx context.Context, item *models.SourceSite) error {
		path := "dnew.php"
		if page > 1 {
			path = fmt.Sprintf("dnew.php?page=%d", page)
		}
		html, err := client.fetchHTML(runCtx, item, resolveMaybeURL(item.BaseURL, path))
		if err != nil {
			return err
		}
		parsed, err := client.parser.ParseListPage(item.BaseURL, html)
		if err != nil {
			return err
		}
		if parsed.CurrentPage == 0 {
			parsed.CurrentPage = page
		}
		result = parsed
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return source, result, nil
}

func (client *SourceClient) FetchComicDetailBundle(ctx context.Context, preferredSourceID *int64, comicID int64) (*models.SourceSite, *models.RemoteComicDetailBundle, error) {
	var result *models.RemoteComicDetailBundle
	source, err := client.withSource(ctx, preferredSourceID, fmt.Sprintf("sync.comic.%d", comicID), func(runCtx context.Context, item *models.SourceSite) error {
		detailHTML, err := client.fetchHTML(runCtx, item, resolveMaybeURL(item.BaseURL, fmt.Sprintf("post.php?ID=%d", comicID)))
		if err != nil {
			return err
		}
		bundle, err := client.parser.ParseDetailPage(item.BaseURL, comicID, detailHTML)
		if err != nil {
			return err
		}
		readerHTML, err := client.fetchHTML(runCtx, item, resolveMaybeURL(item.BaseURL, fmt.Sprintf("readOnline2.php?ID=%d&host_id=0", comicID)))
		if err != nil {
			return err
		}
		if err := client.parser.ParseReaderPage(item.BaseURL, comicID, readerHTML, bundle); err != nil {
			return err
		}
		result = bundle
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return source, result, nil
}

func (client *SourceClient) withSource(ctx context.Context, preferredSourceID *int64, purpose string, fn func(context.Context, *models.SourceSite) error) (*models.SourceSite, error) {
	candidates, err := client.loadCandidates(preferredSourceID)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no source configured")
	}

	errorsSeen := make([]string, 0, len(candidates))
	for index := range candidates {
		item := &candidates[index]
		client.events.Publish("source.request.start", map[string]any{
			"source_id": item.ID,
			"name":      item.Name,
			"purpose":   purpose,
		})
		if err := fn(ctx, item); err != nil {
			errorsSeen = append(errorsSeen, fmt.Sprintf("%s: %v", item.Name, err))
			client.events.Publish("source.request.switch", map[string]any{
				"source_id": item.ID,
				"name":      item.Name,
				"purpose":   purpose,
				"error":     err.Error(),
			})
			continue
		}
		return item, nil
	}
	return nil, fmt.Errorf(strings.Join(errorsSeen, "; "))
}

func (client *SourceClient) loadCandidates(preferredSourceID *int64) ([]models.SourceSite, error) {
	items, err := client.store.List()
	if err != nil {
		return nil, err
	}
	candidates := make([]models.SourceSite, 0, len(items))
	for _, item := range items {
		if !item.Enabled {
			continue
		}
		candidates = append(candidates, item)
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	sort.SliceStable(candidates, func(left int, right int) bool {
		leftUnavailable := strings.EqualFold(candidates[left].Status, "unavailable")
		rightUnavailable := strings.EqualFold(candidates[right].Status, "unavailable")
		if leftUnavailable != rightUnavailable {
			return !leftUnavailable
		}
		if candidates[left].Priority != candidates[right].Priority {
			return candidates[left].Priority < candidates[right].Priority
		}
		return candidates[left].ID < candidates[right].ID
	})
	if preferredSourceID == nil {
		return candidates, nil
	}
	preferred := make([]models.SourceSite, 0, 1)
	others := make([]models.SourceSite, 0, len(candidates))
	for _, item := range candidates {
		if item.ID == *preferredSourceID {
			preferred = append(preferred, item)
			continue
		}
		others = append(others, item)
	}
	return append(preferred, others...), nil
}

func (client *SourceClient) fetchHTML(ctx context.Context, item *models.SourceSite, absoluteURL string) (string, error) {
	settingsMap, err := client.settings.List()
	if err != nil {
		return "", err
	}
	attemptTimeout := parseDurationSeconds(settingsMap["source_request_timeout_seconds"].ValueJSON, 15)
	retries := parseInt(settingsMap["source_request_retries"].ValueJSON, 3)
	failureThreshold := parseInt(settingsMap["source_failure_threshold"].ValueJSON, 3)
	if retries <= 0 {
		retries = 3
	}

	var lastErr error
	for attempt := 1; attempt <= retries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
		request, err := http.NewRequestWithContext(attemptCtx, http.MethodGet, absoluteURL, nil)
		if err != nil {
			cancel()
			return "", err
		}
		request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

		startedAt := time.Now()
		response, requestErr := client.httpClient.Do(request)
		if requestErr == nil && response != nil {
			body, readErr := io.ReadAll(io.LimitReader(response.Body, 8<<20))
			_ = response.Body.Close()
			cancel()
			if readErr == nil && response.StatusCode >= 200 && response.StatusCode < 400 {
				latency := int(time.Since(startedAt).Milliseconds())
				_ = client.store.Update(item.ID, map[string]any{
					"status":               "available",
					"consecutive_failures": 0,
					"last_latency_ms":      latency,
					"last_checked_at":      time.Now(),
					"last_error":           "",
				})
				item.Status = "available"
				item.ConsecutiveFailures = 0
				item.LastError = ""
				client.events.Publish("source.request.ok", map[string]any{
					"source_id":  item.ID,
					"name":       item.Name,
					"url":        absoluteURL,
					"attempt":    attempt,
					"latency_ms": latency,
				})
				return string(body), nil
			}
			if readErr != nil {
				lastErr = readErr
			} else {
				lastErr = fmt.Errorf("unexpected status %d", response.StatusCode)
			}
		} else {
			cancel()
			lastErr = requestErr
		}
		message := "request failed"
		if lastErr != nil {
			message = lastErr.Error()
		}
		eventType := "source.request.retry"
		if ctx.Err() == context.DeadlineExceeded || strings.Contains(strings.ToLower(message), "deadline exceeded") {
			eventType = "source.request.timeout"
		}
		client.events.Publish(eventType, map[string]any{
			"source_id": item.ID,
			"name":      item.Name,
			"url":       absoluteURL,
			"attempt":   attempt,
			"error":     message,
		})
	}

	item.ConsecutiveFailures++
	status := item.Status
	if status == "" {
		status = "unknown"
	}
	if item.ConsecutiveFailures >= failureThreshold {
		status = "unavailable"
	} else {
		status = "degraded"
	}
	message := "request failed"
	if lastErr != nil {
		message = lastErr.Error()
	}
	_ = client.store.Update(item.ID, map[string]any{
		"status":               status,
		"consecutive_failures": item.ConsecutiveFailures,
		"last_checked_at":      time.Now(),
		"last_error":           message,
	})
	item.Status = status
	item.LastError = message
	client.events.Publish("source.request.failed", map[string]any{
		"source_id": item.ID,
		"name":      item.Name,
		"url":       absoluteURL,
		"error":     message,
	})
	return "", fmt.Errorf(message)
}
