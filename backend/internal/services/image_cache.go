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

func (service *ImageProxyService) EnsureCoverLocal(ctx context.Context, comicID int64) (string, error) {
	target, err := service.comicStore.GetCoverTarget(comicID)
	if err != nil {
		return "", err
	}

	if candidate, ok := service.resolveLocalFile(target.LocalRelPath); ok {
		if err := service.comicStore.MarkCoverCached(comicID, target.LocalRelPath); err != nil {
			return "", err
		}
		return candidate, nil
	}
	if strings.TrimSpace(target.CoverURL) == "" {
		return "", fmt.Errorf("cover url missing for comic %d", comicID)
	}

	relPath, absPath := service.buildCachePath("covers", target.CoverURL, "")
	if fileExists(absPath) {
		if err := service.comicStore.MarkCoverCached(comicID, relPath); err != nil {
			return "", err
		}
		return absPath, nil
	}

	value, err, _ := service.group.Do("cover:"+target.CoverURL, func() (interface{}, error) {
		if fileExists(absPath) {
			if err := service.comicStore.MarkCoverCached(comicID, relPath); err != nil {
				return nil, err
			}
			return absPath, nil
		}
		if _, err := service.downloadAndStore(ctx, target.CoverURL, absPath, "image/avif,image/webp,image/apng,image/*,*/*;q=0.8"); err != nil {
			service.events.Publish("cover.cache.failed", map[string]any{"comic_id": comicID, "url": target.CoverURL, "error": err.Error()})
			return nil, err
		}
		if err := service.comicStore.MarkCoverCached(comicID, relPath); err != nil {
			return nil, err
		}
		service.events.Publish("cover.cache.completed", map[string]any{
			"comic_id":   comicID,
			"url":        target.CoverURL,
			"local_path": relPath,
			"file_size":  fileSize(absPath),
		})
		return absPath, nil
	})
	if err != nil {
		return "", err
	}

	resolved, _ := value.(string)
	if resolved == "" {
		return "", fmt.Errorf("cover cache path is empty")
	}
	if err := service.comicStore.MarkCoverCached(comicID, relPath); err != nil {
		return "", err
	}
	return resolved, nil
}

func (service *ImageProxyService) EnsureImageLocal(ctx context.Context, comicID int64, sort int) (string, error) {
	target, err := service.comicStore.GetImageTarget(comicID, sort)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(target.ImageURL) == "" {
		return "", fmt.Errorf("image url missing for comic %d page %d", comicID, sort)
	}

	if candidate, ok := service.resolveLocalFile(target.LocalRelPath); ok {
		if err := service.comicStore.MarkImageCached(comicID, sort, target.LocalRelPath, fileSize(candidate)); err != nil {
			return "", err
		}
		return candidate, nil
	}

	relPath, absPath := service.buildCachePath("images", target.ImageURL, target.Extension)
	if fileExists(absPath) {
		if err := service.comicStore.MarkImageCached(comicID, sort, relPath, fileSize(absPath)); err != nil {
			return "", err
		}
		return absPath, nil
	}

	value, err, _ := service.group.Do("image:"+target.ImageURL, func() (interface{}, error) {
		if fileExists(absPath) {
			if err := service.comicStore.MarkImageCached(comicID, sort, relPath, fileSize(absPath)); err != nil {
				return nil, err
			}
			return absPath, nil
		}
		_, err := service.downloadAndStore(ctx, target.ImageURL, absPath, "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
		if err != nil {
			service.events.Publish("image.cache.failed", map[string]any{"comic_id": comicID, "sort": sort, "url": target.ImageURL, "error": err.Error()})
			return nil, err
		}
		return absPath, nil
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
	service.events.Publish("image.cache.completed", map[string]any{
		"comic_id":   comicID,
		"sort":       sort,
		"url":        target.ImageURL,
		"local_path": relPath,
		"file_size":  fileSize(resolved),
	})
	return resolved, nil
}

func (service *ImageProxyService) downloadAndStore(ctx context.Context, remoteURL string, absPath string, accept string) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return 0, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, remoteURL, nil)
	if err != nil {
		return 0, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	if strings.TrimSpace(accept) != "" {
		request.Header.Set("Accept", accept)
	}

	response, err := service.httpClient.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return 0, fmt.Errorf("unexpected remote status %d", response.StatusCode)
	}

	temporaryFile, err := os.CreateTemp(filepath.Dir(absPath), "asset-*.tmp")
	if err != nil {
		return 0, err
	}
	temporaryName := temporaryFile.Name()
	copied, copyErr := io.Copy(temporaryFile, response.Body)
	closeErr := temporaryFile.Close()
	if copyErr != nil {
		_ = os.Remove(temporaryName)
		return 0, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(temporaryName)
		return 0, closeErr
	}
	if copied <= 0 {
		_ = os.Remove(temporaryName)
		return 0, fmt.Errorf("empty remote body")
	}
	if err := os.Rename(temporaryName, absPath); err != nil {
		_ = os.Remove(temporaryName)
		return 0, err
	}
	return copied, nil
}

func (service *ImageProxyService) resolveLocalFile(localRelPath string) (string, bool) {
	if strings.TrimSpace(localRelPath) == "" {
		return "", false
	}
	candidate := filepath.Join(service.dataDir, filepath.FromSlash(localRelPath))
	if !fileExists(candidate) {
		return "", false
	}
	return candidate, true
}

func (service *ImageProxyService) buildCachePath(kind string, remoteURL string, extension string) (string, string) {
	ext := sanitizeImageExtension(remoteURL, extension)
	hashBytes := sha1.Sum([]byte(strings.TrimSpace(remoteURL)))
	hash := hex.EncodeToString(hashBytes[:])
	relPath := filepath.ToSlash(filepath.Join("cache", kind, hash[:2], hash+"."+ext))
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
