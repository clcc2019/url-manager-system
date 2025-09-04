package routes

import (
	"url-manager-system/backend/internal/api/handlers"
	"url-manager-system/backend/internal/api/middleware"
	"url-manager-system/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(serviceContainer *services.Container) *gin.Engine {
	router := gin.New()

	// 全局中间件
	router.Use(middleware.SetupLogger())
	router.Use(middleware.Recovery())
	router.Use(middleware.SetupCORS())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RequestSize(10 << 20)) // 10MB

	// 对API路由添加速率限制
	router.Use(middleware.RateLimiter())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "url-manager-system",
		})
	})

	// API路由组
	api := router.Group("/api/v1")
	{
		setupProjectRoutes(api, serviceContainer)
		setupURLRoutes(api, serviceContainer)
	}

	return router
}

// setupProjectRoutes 设置项目路由
func setupProjectRoutes(api *gin.RouterGroup, serviceContainer *services.Container) {
	projectHandler := handlers.NewProjectHandler(serviceContainer.ProjectService)

	projects := api.Group("/projects")
	{
		projects.POST("", projectHandler.CreateProject)
		projects.GET("", projectHandler.ListProjects)
		projects.GET("/:id", projectHandler.GetProject)
		projects.PUT("/:id", projectHandler.UpdateProject)
		projects.DELETE("/:id", projectHandler.DeleteProject)

		// 项目下的URL管理
		urlHandler := handlers.NewURLHandler(serviceContainer.URLService, serviceContainer.CleanupService)
		projects.POST("/:id/urls", urlHandler.CreateEphemeralURL)
		projects.GET("/:id/urls", urlHandler.ListEphemeralURLs)
	}
}

// setupURLRoutes 设置URL路由
func setupURLRoutes(api *gin.RouterGroup, serviceContainer *services.Container) {
	urlHandler := handlers.NewURLHandler(serviceContainer.URLService, serviceContainer.CleanupService)

	urls := api.Group("/urls")
	{
		urls.GET("/:id", urlHandler.GetEphemeralURL)
		urls.DELETE("/:id", urlHandler.DeleteEphemeralURL)
		// urls.GET("/path/:path", urlHandler.GetURLByPath) // 可选：根据路径查询
	}
}
