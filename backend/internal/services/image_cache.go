package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/huyaaaaaa/hehuan-reader/internal/store"
	"golang.org/x/sync/singleflight"
)

type ImageProxyService struct {
	dataDir    string
	comicStore *store.ComicStore
	events     *EventBroker
	httpClient *http.Client
	group      singleflight.Group
}

func NewImageProxyService(dataDir string, comicStore *store.ComicStore, events *EventBroker) *ImageProxyService {
	return &ImageProxyService{
		dataDir:    dataDir,
		comicStore: comicStore,
		events:     events,
		httpClient: &http.Client{Timeout: 2 * time.Minute},
	}
}

func (service *ImageProxyService) EnsureImageLocal(ctx context.Context, comicID int64, sort int) (string, error) {
	target, err := service.comicStore.GetImageTarget(comicID, sort)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(target.ImageURL) == "" {
		return "", fmt.Errorf("image url missing for comic %d page %d", comicID, sort)
	}

	relPath, absPath := service.buildCachePath(target.ImageURL, target.Extension)
	if target.LocalRelPath != "" {
		candidate := filepath.Join(service.dataDir, filepath.FromSlash(target.LocalRelPath))
		if fileExists(candidate) {
			return candidate, nil
		}
	}

	if fileExists(absPath) {
		if err := service.comicStore.MarkImageCached(comicID, sort, relPath, fileSize(absPath)); err != nil {
			return "", err
		}
		return absPath, nil
	}

	value, err, _ := service.group.Do(target.ImageURL, func() (interface{}, error) {
		if fileExists(absPath) {
			if err := service.comicStore.MarkImageCached(comicID, sort, relPath, fileSize(absPath)); err != nil {
				return nil, err
			}
			return absPath, nil
		}
		return service.downloadAndStore(ctx, comicID, sort, target.ImageURL, relPath, absPath)
	})
	if err != nil {
		return "", err
	}
	resolved, _ := value.(string)
	if resolved == "" {
		return "", fmt.Errorf("image cache path is empty")
	}
	if err := service.comicStore.MarkImageCached(comicID, sort, relPath, fileSize(resolved)); err != nil {
		return "", err
	}
	return resolved, nil
}

func (service *ImageProxyService) downloadAndStore(ctx context.Context, comicID int64, sort int, imageURL string, relPath string, absPath string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	request.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
	timeStarted := time.Now()
	response, err := service.httpClient.Do(request)
	if err != nil {
		service.events.Publish("image.cache.failed", map[string]any{"comic_id": comicID, "sort": sort, "url": imageURL, "error": err.Error()})
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		err = fmt.Errorf("unexpected image status %d", response.StatusCode)
		service.events.Publish("image.cache.failed", map[string]any{"comic_id": comicID, "sort": sort, "url": imageURL, "error": err.Error()})
		return "", err
	}

	temporaryFile, err := os.CreateTemp(filepath.Dir(absPath), "img-*.tmp")
	if err != nil {
		return "", err
	}
	temporaryName := temporaryFile.Name()
	copied, copyErr := io.Copy(temporaryFile, response.Body)
	closeErr := temporaryFile.Close()
	if copyErr != nil {
		_ = os.Remove(temporaryName)
		service.events.Publish("image.cache.failed", map[string]any{"comic_id": comicID, "sort": sort, "url": imageURL, "error": copyErr.Error()})
		return "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(temporaryName)
		return "", closeErr
	}
	if copied <= 0 {
		_ = os.Remove(temporaryName)
		return "", fmt.Errorf("empty image body")
	}
	if err := os.Rename(temporaryName, absPath); err != nil {
		_ = os.Remove(temporaryName)
		return "", err
	}
	if err := service.comicStore.MarkImageCached(comicID, sort, relPath, copied); err != nil {
		return "", err
	}
	service.events.Publish("image.cache.completed", map[string]any{
		"comic_id":   comicID,
		"sort":       sort,
		"url":        imageURL,
		"local_path": relPath,
		"file_size":  copied,
		"elapsed_ms": time.Since(timeStarted).Milliseconds(),
	})
	return absPath, nil
}

func (service *ImageProxyService) buildCachePath(imageURL string, extension string) (string, string) {
	ext := sanitizeImageExtension(imageURL, extension)
	hashBytes := sha1.Sum([]byte(strings.TrimSpace(imageURL)))
	hash := hex.EncodeToString(hashBytes[:])
	relPath := filepath.ToSlash(filepath.Join("cache", "images", hash[:2], hash+"."+ext))
	absPath := filepath.Join(service.dataDir, filepath.FromSlash(relPath))
	return relPath, absPath
}

func sanitizeImageExtension(imageURL string, extension string) string {
	ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(extension, ".")))
	if ext == "" {
		if parsed, err := url.Parse(strings.TrimSpace(imageURL)); err == nil {
			pathExt := strings.TrimPrefix(strings.ToLower(filepath.Ext(parsed.Path)), ".")
			ext = pathExt
		}
	}
	if ext == "jpeg" {
		ext = "jpg"
	}
	if ext == "" {
		return "jpg"
	}
	for _, char := range ext {
		if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9') {
			return "jpg"
		}
	}
	if len(ext) > 8 {
		return "jpg"
	}
	return ext
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
