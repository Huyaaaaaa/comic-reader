package main

import (
	"comic-viewer-claude/internal/api"
	"comic-viewer-claude/internal/cache"
	"comic-viewer-claude/internal/crawler"
	"comic-viewer-claude/internal/download"
	"comic-viewer-claude/internal/event"
	"comic-viewer-claude/internal/middleware"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/internal/service"
	"comic-viewer-claude/internal/source"
	"comic-viewer-claude/pkg/config"
	"comic-viewer-claude/pkg/logger"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(
		cfg.Log.Level,
		cfg.Log.File,
		cfg.Log.MaxSize,
		cfg.Log.MaxBackups,
		cfg.Log.MaxAge,
	); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("启动漫画浏览器服务", zap.String("version", "1.0.0"))

	// 初始化数据库
	repo, err := repository.New(cfg.Database.Path)
	if err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	defer repo.Close()
	logger.Info("数据库初始化成功", zap.String("path", cfg.Database.Path))

	// 初始化爬虫客户端
	crawlerClient := crawler.NewClient(&cfg.Crawler)
	logger.Info("爬虫客户端初始化成功")

	// 初始化封面缓存管理器
	coverCache := cache.NewCoverCacheManager(
		repo,
		crawlerClient,
		cfg.Cache.CoverMax,
		cfg.Cache.CoverTargetSizeKB,
	)
	if cfg.Cache.CoverStrategy == "auto" {
		coverCache.Start()
		defer coverCache.Stop()
	}

	// 执行数据迁移
	repository.RunMigration(repo.GetDB())

	// 初始化下载引擎
	dlEngine := download.NewEngine(repo, crawlerClient, &cfg.Download, &cfg.Crawler)
	dlEngine.Start()
	defer dlEngine.Stop()

	// 初始化服务
	comicService := service.NewComicService(repo, crawlerClient, coverCache)
	searchService := service.NewSearchService(repo, comicService)
	settingsService := service.NewSettingsService(repo)
	downloadService := service.NewDownloadService(repo, dlEngine, &cfg.Download)
	ieService := service.NewImportExportService(repo, &cfg.Download)
	updateService := service.NewUpdateService(repo, comicService)

	// 初始化源站管理器
	sourceManager := source.NewManager(repo)
	if err := sourceManager.Load(); err != nil {
		logger.Warn("加载源站列表失败", zap.Error(err))
	}

	// 初始化 SSE 事件服务
	eventService := event.NewService()

	// 启动定时更新网站总数任务
	stopStats := make(chan struct{})
	go comicService.StartSiteStatsUpdater(stopStats)

	// 启动源站健康检查
	stopHealth := make(chan struct{})
	go sourceManager.StartHealthChecker(stopHealth)

	// 初始化处理器
	comicHandler := api.NewComicHandler(comicService)
	v2Handler := api.NewV2Handler(searchService, settingsService, comicService, repo)
	downloadHandler := api.NewDownloadHandler(downloadService)
	ieHandler := api.NewImportExportHandler(ieService)
	updateHandler := api.NewUpdateHandler(updateService)
	sourceHandler := api.NewSourceHandler(sourceManager)
	eventHandler := api.NewEventHandler(eventService)
	imageHandler := api.NewImageHandler(repo, crawlerClient)
	metaHandler := api.NewMetaHandler(repo)

	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎
	engine := gin.New()
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS(cfg.Server.CorsOrigins))
	// 设置路由
	router := api.NewRouter(comicHandler, v2Handler, downloadHandler, ieHandler, updateHandler, sourceHandler, eventHandler, imageHandler, metaHandler)
	router.Setup(engine)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("服务器启动", zap.String("addr", addr), zap.String("mode", cfg.Server.Mode))

	// 优雅关闭
	go func() {
		if err := engine.Run(addr); err != nil {
			logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	close(stopStats)
	close(stopHealth)
	logger.Info("服务器正在关闭...")
}
