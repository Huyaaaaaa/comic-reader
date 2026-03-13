package api

import (
	"github.com/gin-gonic/gin"
)

// Router 路由器
type Router struct {
	comicHandler    *ComicHandler
	v2Handler       *V2Handler
	downloadHandler *DownloadHandler
}

// NewRouter 创建路由器
func NewRouter(comicHandler *ComicHandler, v2Handler *V2Handler, downloadHandler *DownloadHandler) *Router {
	return &Router{
		comicHandler:    comicHandler,
		v2Handler:       v2Handler,
		downloadHandler: downloadHandler,
	}
}

// Setup 设置路由
func (r *Router) Setup(engine *gin.Engine) {
	// API 路由组
	api := engine.Group("/api")
	{
		// 漫画相关
		comics := api.Group("/comics")
		{
			comics.GET("", r.comicHandler.GetList)           // 获取列表
			comics.GET("/:id", r.comicHandler.GetDetail)     // 获取详情
			comics.GET("/:id/images", r.comicHandler.GetReaderImages) // 获取图片列表
			comics.GET("/search", r.comicHandler.Search)     // 搜索
			comics.GET("/filter", r.comicHandler.Filter)     // 筛选
			comics.POST("/:id/history", r.comicHandler.AddReadingHistory) // 添加历史
			comics.POST("/:id/favorite", r.comicHandler.ToggleFavorite)   // 切换收藏
		}

		// 收藏
		api.GET("/favorites", r.comicHandler.GetFavorites)

		// 历史记录
		api.GET("/history", r.comicHandler.GetHistory)

		// 仪表盘
		dashboard := api.Group("/dashboard")
		{
			dashboard.GET("/stats", r.comicHandler.GetDashboardStats) // 统计数据
		}

		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// V2 API 路由组
		v2 := api.Group("/v2")
		{
			v2.GET("/search", r.v2Handler.Search)
			v2.GET("/search/history", r.v2Handler.GetSearchHistory)
			v2.DELETE("/search/history", r.v2Handler.ClearSearchHistory)
			v2.GET("/settings", r.v2Handler.GetSettings)
			v2.PUT("/settings", r.v2Handler.UpdateSettings)
			v2.GET("/cache/status", r.v2Handler.GetCacheStatus)
			v2.GET("/history", r.v2Handler.GetHistoryWithProgress)
			v2.PUT("/history/:id/progress", r.v2Handler.UpdateProgress)
			v2.GET("/offline/comics", r.v2Handler.GetOfflineComics)
			v2.GET("/comics/:id/cache", r.v2Handler.GetComicCacheState)

			// 下载管理
			downloads := v2.Group("/downloads")
			{
				downloads.POST("", r.downloadHandler.CreateTask)
				downloads.POST("/precheck", r.downloadHandler.Precheck)
				downloads.GET("", r.downloadHandler.GetTasks)
				downloads.POST("/:id/pause", r.downloadHandler.PauseTask)
				downloads.POST("/:id/resume", r.downloadHandler.ResumeTask)
				downloads.POST("/:id/cancel", r.downloadHandler.CancelTask)
				downloads.POST("/:id/retry", r.downloadHandler.RetryTask)
			}
		}
	}
}
