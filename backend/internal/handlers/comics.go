package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"github.com/huyaaaaaa/hehuan-reader/internal/store"
	"gorm.io/gorm"
)

type ComicsHandler struct {
	store *store.ComicStore
}

func NewComicsHandler(store *store.ComicStore) *ComicsHandler {
	return &ComicsHandler{store: store}
}

func (handler *ComicsHandler) List(c *gin.Context) {
	query := models.ComicQuery{
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", 100),
		Search:   strings.TrimSpace(c.Query("search")),
	}
	query.CategoryID = int64QueryPtr(c, "category_id")
	query.AuthorID = int64QueryPtr(c, "author_id")
	query.TagID = int64QueryPtr(c, "tag_id")
	result, err := handler.store.List(query)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, result, okMeta())
}

func (handler *ComicsHandler) Detail(c *gin.Context) {
	comicID, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	result, err := handler.store.Detail(comicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "comic not found", noContentMeta())
			return
		}
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, result, okMeta())
}

func (handler *ComicsHandler) Images(c *gin.Context) {
	comicID, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	result, err := handler.store.Images(comicID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, result, okMeta())
}

func (handler *ComicsHandler) Search(c *gin.Context) {
	query := models.ComicQuery{
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", 20),
		Search:   strings.TrimSpace(c.Query("keyword")),
	}
	result, err := handler.store.List(query)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, result, Meta{Offline: false, Stale: true, Source: "hybrid", Message: "本地结果优先，后续可补远程搜索"})
}

func (handler *ComicsHandler) Tags(c *gin.Context) {
	tags, err := handler.store.Tags()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, gin.H{"tags": tags}, okMeta())
}

func (handler *ComicsHandler) Categories(c *gin.Context) {
	categories, err := handler.store.Categories()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, gin.H{"categories": categories}, okMeta())
}

func (handler *ComicsHandler) Authors(c *gin.Context) {
	authors, total, err := handler.store.Authors(intQuery(c, "page", 1), intQuery(c, "page_size", 20), c.Query("search"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, gin.H{"authors": authors, "total": total, "page": intQuery(c, "page", 1), "page_size": intQuery(c, "page_size", 20)}, okMeta())
}

func intQuery(c *gin.Context, key string, fallback int) int {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func int64QueryPtr(c *gin.Context, key string) *int64 {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &value
}

func pathInt64(c *gin.Context, key string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id", okMeta())
		return 0, false
	}
	return value, true
}
