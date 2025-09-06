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
		// 公开路由（不需要认证）
		setupAuthRoutes(api, serviceContainer)

		// 需要认证的路由
		authorized := api.Group("")
		authorized.Use(middleware.AuthMiddleware(serviceContainer.AuthService))
		{
			setupProjectRoutes(authorized, serviceContainer)
			setupURLRoutes(authorized, serviceContainer)
			setupTemplateRoutes(authorized, serviceContainer)
			setupUserRoutes(authorized, serviceContainer)
		}
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
		projects.POST("/:id/urls/from-template", urlHandler.CreateEphemeralURLFromTemplate)
		projects.GET("/:id/urls", urlHandler.ListEphemeralURLs)

		// 项目统计
		projects.GET("/stats", projectHandler.GetProjectStats)
	}
}

// setupURLRoutes 设置URL路由
func setupURLRoutes(api *gin.RouterGroup, serviceContainer *services.Container) {
	urlHandler := handlers.NewURLHandler(serviceContainer.URLService, serviceContainer.CleanupService)

	urls := api.Group("/urls")
	{
		urls.GET("/:id", urlHandler.GetEphemeralURL)
		urls.PUT("/:id", urlHandler.UpdateEphemeralURL)
		urls.DELETE("/:id", urlHandler.DeleteEphemeralURL)
		urls.POST("/:id/deploy", urlHandler.DeployURL)
		urls.POST("/validate-cleanup", urlHandler.ValidateAndCleanupData)

		// 容器状态、事件和日志相关API
		urls.GET("/:id/containers/status", urlHandler.GetURLContainerStatus)
		urls.GET("/:id/events", urlHandler.GetURLPodEvents)
		urls.GET("/:id/logs", urlHandler.GetURLContainerLogs)

		// urls.GET("/path/:path", urlHandler.GetURLByPath) // 可选：根据路径查询
	}
}

// setupTemplateRoutes 设置模版路由
func setupTemplateRoutes(api *gin.RouterGroup, serviceContainer *services.Container) {
	templateHandler := handlers.NewTemplateHandler(serviceContainer.TemplateService)

	templates := api.Group("/templates")
	{
		templates.POST("", templateHandler.CreateTemplate)
		templates.GET("", templateHandler.ListTemplates)
		templates.GET("/:id", templateHandler.GetTemplate)
		templates.PUT("/:id", templateHandler.UpdateTemplate)
		templates.DELETE("/:id", templateHandler.DeleteTemplate)
		templates.GET("/:id/variables", templateHandler.GetTemplateVariables)
		templates.POST("/:id/preview", templateHandler.PreviewTemplate)
	}
}

// setupAuthRoutes 设置认证路由（不需要认证）
func setupAuthRoutes(api *gin.RouterGroup, serviceContainer *services.Container) {
	authHandler := handlers.NewAuthHandler(serviceContainer.AuthService)

	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout) // 前端处理，后端只返回成功
	}
}

// setupUserRoutes 设置用户路由（需要认证）
func setupUserRoutes(api *gin.RouterGroup, serviceContainer *services.Container) {
	authHandler := handlers.NewAuthHandler(serviceContainer.AuthService)

	users := api.Group("/users")
	{
		// 用户信息相关
		users.GET("/profile", authHandler.GetProfile)
		users.PUT("/password", authHandler.ChangePassword)

		// 管理员功能（所有登录用户都可以访问）
		users.POST("/register", authHandler.Register)
		users.GET("", authHandler.ListUsers)
	}
}
