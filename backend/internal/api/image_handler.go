package api

import (
	"comic-viewer-claude/internal/crawler"
	"comic-viewer-claude/internal/repository"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"
)

// ImageHandler 图片代理处理器
type ImageHandler struct {
	repo    *repository.Repository
	crawler *crawler.Client
	sf      singleflight.Group
}

// NewImageHandler 创建图片代理处理器
func NewImageHandler(repo *repository.Repository, crawler *crawler.Client) *ImageHandler {
	return &ImageHandler{
		repo:    repo,
		crawler: crawler,
	}
}

// ProxyImage GET /api/images/proxy?comic_id=X&sort=Y&url=Z
func (h *ImageHandler) ProxyImage(c *gin.Context) {
	comicIDStr := c.Query("comic_id")
	sortStr := c.Query("sort")
	rawURL := c.Query("url")

	if rawURL == "" {
		Error(c, http.StatusBadRequest, CodeBadRequest, "缺少 url 参数")
		return
	}

	comicID, _ := strconv.Atoi(comicIDStr)
	sort, _ := strconv.Atoi(sortStr)

	// 如果有 comic_id 和 sort，先查本地文件
	if comicID > 0 && sort > 0 {
		images, err := h.repo.GetComicImages(comicID)
		if err == nil {
			for _, img := range images {
				if img.Sort == sort && img.LocalPath != "" {
					if _, err := os.Stat(img.LocalPath); err == nil {
						c.File(img.LocalPath)
						return
					}
				}
			}
		}
	}

	// singleflight 去重：相同 URL 只下载一次
	key := fmt.Sprintf("img:%s", rawURL)
	data, err, _ := h.sf.Do(key, func() (interface{}, error) {
		return h.crawler.DownloadImage(rawURL)
	})
	if err != nil {
		Error(c, http.StatusBadGateway, CodeInternal, "图片下载失败: "+err.Error())
		return
	}

	imgBytes := data.([]byte)
	contentType := http.DetectContentType(imgBytes)
	c.Data(http.StatusOK, contentType, imgBytes)
}
