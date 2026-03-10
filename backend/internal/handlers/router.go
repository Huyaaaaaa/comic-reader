package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/huyaaaaaa/hehuan-reader/internal/config"
	"github.com/huyaaaaaa/hehuan-reader/internal/services"
	"github.com/huyaaaaaa/hehuan-reader/internal/store"
)

func NewRouter(
	cfg config.Config,
	settingsStore *store.SettingsStore,
	sourceStore *store.SourceStore,
	comicStore *store.ComicStore,
	libraryStore *store.LibraryStore,
	sourceHealth *services.SourceHealthService,
	syncService *services.SyncService,
	imageProxyService *services.ImageProxyService,
	events *services.EventBroker,
) *gin.Engine {
	engine := gin.Default()
	engine.Use(cors(cfg.AllowedOrigins))

	healthHandler := NewHealthHandler(cfg)
	settingsHandler := NewSettingsHandler(settingsStore, events)
	sourcesHandler := NewSourcesHandler(sourceStore, sourceHealth, events)
	eventsHandler := NewEventsHandler(events)
	comicsHandler := NewComicsHandler(comicStore)
	libraryHandler := NewLibraryHandler(libraryStore)
	syncHandler := NewSyncHandler(syncService)
	mediaHandler := NewMediaHandler(imageProxyService)

	api := engine.Group("/api")
	{
		api.GET("/health", healthHandler.Health)
		api.GET("/settings", settingsHandler.List)
		api.PUT("/settings/:key", settingsHandler.Update)
		api.GET("/sources", sourcesHandler.List)
		api.POST("/sources", sourcesHandler.Create)
		api.PUT("/sources/:id", sourcesHandler.Update)
		api.DELETE("/sources/:id", sourcesHandler.Delete)
		api.POST("/sources/:id/check", sourcesHandler.Check)
		api.GET("/events/stream", eventsHandler.Stream)
		api.POST("/sync/head", syncHandler.SyncHead)
		api.POST("/sync/comics/:id/detail", syncHandler.SyncComicDetail)
		api.GET("/covers/proxy", mediaHandler.ProxyCover)
		api.GET("/images/proxy", mediaHandler.ProxyImage)
		api.GET("/comics", comicsHandler.List)
		api.GET("/comics/:id", comicsHandler.Detail)
		api.GET("/comics/:id/images", comicsHandler.Images)
		api.GET("/search", comicsHandler.Search)
		api.GET("/tags", comicsHandler.Tags)
		api.GET("/categories", comicsHandler.Categories)
		api.GET("/authors", comicsHandler.Authors)
		api.POST("/favorites", libraryHandler.CreateFavorite)
		api.DELETE("/favorites/:comic_id", libraryHandler.DeleteFavorite)
		api.GET("/favorites", libraryHandler.ListFavorites)
		api.POST("/history", libraryHandler.UpsertHistory)
		api.GET("/history", libraryHandler.ListHistory)
		api.POST("/search/history", libraryHandler.AddSearchHistory)
		api.GET("/search/history", libraryHandler.ListSearchHistory)
		api.DELETE("/search/history", libraryHandler.ClearSearchHistory)
	}

	return engine
}

func cors(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[strings.TrimSpace(origin)] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
