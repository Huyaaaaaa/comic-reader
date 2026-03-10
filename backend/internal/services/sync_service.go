package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"github.com/huyaaaaaa/hehuan-reader/internal/store"
)

type SyncService struct {
	settings *store.SettingsStore
	syncs    *store.SyncStore
	sources  *SourceClient
	events   *EventBroker
}

func NewSyncService(settingsStore *store.SettingsStore, syncStore *store.SyncStore, sourceClient *SourceClient, events *EventBroker) *SyncService {
	return &SyncService{settings: settingsStore, syncs: syncStore, sources: sourceClient, events: events}
}

func (service *SyncService) SyncHead(ctx context.Context, preferredSourceID *int64, pages int) (*models.SyncHeadResult, error) {
	settingsMap, err := service.settings.List()
	if err != nil {
		return nil, err
	}
	if pages <= 0 {
		pages = parseInt(settingsMap["sync_head_scan_pages"].ValueJSON, 5)
	}
	if pages <= 0 {
		pages = 5
	}

	paramsJSON, _ := json.Marshal(map[string]any{"source_id": preferredSourceID, "pages": pages})
	jobID, jobErr := service.syncs.CreateJob("head_scan", string(paramsJSON))
	jobSucceeded := false
	if jobErr == nil {
		defer func() {
			status := "failed"
			progress := 0.0
			if jobSucceeded {
				status = "completed"
				progress = 1
			}
			_ = service.syncs.UpdateJob(jobID, status, progress, "")
		}()
	}

	service.events.Publish("sync.head.started", map[string]any{"pages": pages, "source_id": preferredSourceID})
	capturedAt := time.Now()
	var selectedSource *models.SourceSite
	var selectedSourceID *int64
	var firstPage *models.RemoteListPage
	var totalPages int
	scannedPages := 0
	scannedItems := 0
	updatedItems := 0
	lastPageCount := 0

	for pageIndex := 1; pageIndex <= pages; pageIndex++ {
		source, listPage, err := service.sources.FetchHeadPage(ctx, selectedSourceID, pageIndex)
		if err != nil {
			if jobErr == nil {
				_ = service.syncs.UpdateJob(jobID, "failed", 0, err.Error())
			}
			service.events.Publish("sync.head.failed", map[string]any{"page": pageIndex, "error": err.Error()})
			return nil, err
		}
		if selectedSource == nil {
			selectedSource = source
			selectedSourceID = &source.ID
		}
		if firstPage == nil {
			firstPage = listPage
			totalPages = listPage.TotalPages
		}
		currentUpdated, err := service.syncs.UpsertHeadPage(source, *listPage)
		if err != nil {
			if jobErr == nil {
				_ = service.syncs.UpdateJob(jobID, "failed", 0, err.Error())
			}
			service.events.Publish("sync.head.failed", map[string]any{"page": pageIndex, "error": err.Error()})
			return nil, err
		}
		scannedPages++
		scannedItems += len(listPage.Items)
		updatedItems += currentUpdated
		if totalPages > 0 && pageIndex >= totalPages {
			lastPageCount = len(listPage.Items)
		}
		if jobErr == nil {
			progress := float64(scannedPages) / float64(maxInt(pages, 1))
			_ = service.syncs.UpdateJob(jobID, "running", progress, "")
		}
		service.events.Publish("sync.head.page", map[string]any{
			"page":          pageIndex,
			"items":         len(listPage.Items),
			"total_pages":   totalPages,
			"updated_items": currentUpdated,
		})
		if lastPageCount > 0 {
			break
		}
	}

	if totalPages > 0 && firstPage != nil && lastPageCount == 0 {
		source, lastPage, err := service.sources.FetchHeadPage(ctx, selectedSourceID, totalPages)
		if err != nil {
			if jobErr == nil {
				_ = service.syncs.UpdateJob(jobID, "failed", 0, err.Error())
			}
			service.events.Publish("sync.head.failed", map[string]any{"page": totalPages, "error": err.Error()})
			return nil, err
		}
		if selectedSource == nil {
			selectedSource = source
			selectedSourceID = &source.ID
		}
		lastPageCount = len(lastPage.Items)
	}

	totalComics := 0
	if firstPage != nil {
		switch {
		case totalPages <= 1:
			totalComics = len(firstPage.Items)
		case len(firstPage.Items) > 0 && lastPageCount > 0:
			totalComics = (totalPages-1)*len(firstPage.Items) + lastPageCount
		}
	}

	if selectedSource != nil {
		if err := service.syncs.SaveCatalogSnapshot(selectedSource.ID, totalComics, totalPages, lastPageCount); err != nil {
			if jobErr == nil {
				_ = service.syncs.UpdateJob(jobID, "failed", 0, err.Error())
			}
			return nil, err
		}
	}

	result := &models.SyncHeadResult{
		ScannedPages:  scannedPages,
		TotalPages:    totalPages,
		ScannedItems:  scannedItems,
		UpdatedItems:  updatedItems,
		TotalComics:   totalComics,
		LastPageCount: lastPageCount,
		CapturedAt:    capturedAt,
	}
	if selectedSource != nil {
		result.SourceID = selectedSource.ID
		result.SourceName = selectedSource.Name
	}
	service.events.Publish("sync.head.completed", map[string]any{
		"source_id":       result.SourceID,
		"source_name":     result.SourceName,
		"scanned_pages":   result.ScannedPages,
		"total_pages":     result.TotalPages,
		"scanned_items":   result.ScannedItems,
		"updated_items":   result.UpdatedItems,
		"total_comics":    result.TotalComics,
		"last_page_count": result.LastPageCount,
	})
	jobSucceeded = true
	return result, nil
}

func (service *SyncService) SyncComicDetail(ctx context.Context, preferredSourceID *int64, comicID int64) (*models.SyncComicResult, error) {
	paramsJSON, _ := json.Marshal(map[string]any{"source_id": preferredSourceID, "comic_id": comicID})
	jobID, jobErr := service.syncs.CreateJob("comic_detail_refresh", string(paramsJSON))
	jobSucceeded := false
	if jobErr == nil {
		defer func() {
			status := "failed"
			progress := 0.0
			if jobSucceeded {
				status = "completed"
				progress = 1
			}
			_ = service.syncs.UpdateJob(jobID, status, progress, "")
		}()
	}

	service.events.Publish("sync.comic.started", map[string]any{"comic_id": comicID, "source_id": preferredSourceID})
	source, bundle, err := service.sources.FetchComicDetailBundle(ctx, preferredSourceID, comicID)
	if err != nil {
		if jobErr == nil {
			_ = service.syncs.UpdateJob(jobID, "failed", 0, err.Error())
		}
		service.events.Publish("sync.comic.failed", map[string]any{"comic_id": comicID, "error": err.Error()})
		return nil, err
	}
	if err := service.syncs.UpsertComicDetail(*bundle); err != nil {
		if jobErr == nil {
			_ = service.syncs.UpdateJob(jobID, "failed", 0, err.Error())
		}
		service.events.Publish("sync.comic.failed", map[string]any{"comic_id": comicID, "error": err.Error()})
		return nil, err
	}
	result := &models.SyncComicResult{
		ComicID:      comicID,
		Title:        bundle.Title,
		ImagesTotal:  len(bundle.Images),
		AuthorsTotal: len(bundle.Authors),
		TagsTotal:    len(bundle.Tags),
		CapturedAt:   time.Now(),
	}
	if source != nil {
		result.SourceID = source.ID
		result.SourceName = source.Name
	}
	service.events.Publish("sync.comic.completed", map[string]any{
		"comic_id":      comicID,
		"source_id":     result.SourceID,
		"source_name":   result.SourceName,
		"title":         result.Title,
		"images_total":  result.ImagesTotal,
		"authors_total": result.AuthorsTotal,
		"tags_total":    result.TagsTotal,
	})
	jobSucceeded = true
	return result, nil
}

func maxInt(value int, minimum int) int {
	if value < minimum {
		return minimum
	}
	return value
}
