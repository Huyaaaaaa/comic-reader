package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huyaaaaaa/hehuan-reader/internal/models"
	"github.com/huyaaaaaa/hehuan-reader/internal/store"
)

type LibraryHandler struct {
	store *store.LibraryStore
}

type favoriteRequest struct {
	ComicID       int64 `json:"comic_id" binding:"required"`
	EnsureOffline bool  `json:"ensure_offline"`
}

type historyRequest struct {
	ComicID int64                 `json:"comic_id" binding:"required"`
	Locator models.ReadingLocator `json:"locator"`
}

type searchHistoryRequest struct {
	Keyword string `json:"keyword" binding:"required"`
}

func NewLibraryHandler(store *store.LibraryStore) *LibraryHandler {
	return &LibraryHandler{store: store}
}

func (handler *LibraryHandler) CreateFavorite(c *gin.Context) {
	var request favoriteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	if err := handler.store.SetFavorite(defaultUserID, request.ComicID, request.EnsureOffline); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusCreated, gin.H{"comic_id": request.ComicID, "ensure_offline": request.EnsureOffline}, okMeta())
}

func (handler *LibraryHandler) DeleteFavorite(c *gin.Context) {
	comicID, ok := pathInt64(c, "comic_id")
	if !ok {
		return
	}
	if err := handler.store.RemoveFavorite(defaultUserID, comicID); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, gin.H{"comic_id": comicID}, okMeta())
}

func (handler *LibraryHandler) ListFavorites(c *gin.Context) {
	comicID := int64QueryPtr(c, "comic_id")
	items, total, err := handler.store.ListFavorites(defaultUserID, comicID, intQuery(c, "page", 1), intQuery(c, "page_size", 20))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, gin.H{"favorites": items, "total": total, "page": intQuery(c, "page", 1), "page_size": intQuery(c, "page_size", 20)}, okMeta())
}

func (handler *LibraryHandler) UpsertHistory(c *gin.Context) {
	var request historyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	if err := handler.store.UpsertHistory(defaultUserID, request.ComicID, request.Locator); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusCreated, gin.H{"comic_id": request.ComicID}, okMeta())
}

func (handler *LibraryHandler) ListHistory(c *gin.Context) {
	items, total, err := handler.store.ListHistory(defaultUserID, intQuery(c, "page", 1), intQuery(c, "page_size", 20))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, gin.H{"history": items, "total": total, "page": intQuery(c, "page", 1), "page_size": intQuery(c, "page_size", 20)}, okMeta())
}

func (handler *LibraryHandler) AddSearchHistory(c *gin.Context) {
	var request searchHistoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), okMeta())
		return
	}
	if err := handler.store.AddSearchHistory(defaultUserID, request.Keyword); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusCreated, gin.H{"keyword": request.Keyword}, okMeta())
}

func (handler *LibraryHandler) ListSearchHistory(c *gin.Context) {
	items, err := handler.store.ListSearchHistory(defaultUserID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, gin.H{"history": items}, okMeta())
}

func (handler *LibraryHandler) ClearSearchHistory(c *gin.Context) {
	if err := handler.store.ClearSearchHistory(defaultUserID); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error(), okMeta())
		return
	}
	respond(c, http.StatusOK, gin.H{"cleared": true}, okMeta())
}
