package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/huyaaaaaa/hehuan-reader/internal/config"
	"github.com/huyaaaaaa/hehuan-reader/internal/database"
	"github.com/huyaaaaaa/hehuan-reader/internal/handlers"
	"github.com/huyaaaaaa/hehuan-reader/internal/services"
	"github.com/huyaaaaaa/hehuan-reader/internal/store"
)

type App struct {
	cfg        config.Config
	httpServer *http.Server
}

func New() (*App, error) {
	cfg := config.Load()
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	events := services.NewEventBroker()
	settingsStore := store.NewSettingsStore(db)
	sourceStore := store.NewSourceStore(db)
	comicStore := store.NewComicStore(db)
	libraryStore := store.NewLibraryStore(db)
	syncStore := store.NewSyncStore(db)
	sourceHealth := services.NewSourceHealthService(sourceStore, settingsStore, events)
	sourceClient := services.NewSourceClient(sourceStore, settingsStore, events)
	syncService := services.NewSyncService(settingsStore, syncStore, sourceClient, events)
	imageProxyService := services.NewImageProxyService(cfg.DataDir, comicStore, events)

	router := handlers.NewRouter(cfg, settingsStore, sourceStore, comicStore, libraryStore, sourceHealth, syncService, imageProxyService, events)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return &App{cfg: cfg, httpServer: server}, nil
}

func (app *App) Run() error {
	go func() {
		log.Printf("%s listening on %s", app.cfg.AppName, app.cfg.HTTPAddr)
		if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return app.httpServer.Shutdown(ctx)
}
