package main

import (
	"log"

	"GoRestExample/config"
	"GoRestExample/handlers"
	"GoRestExample/middleware"
	"GoRestExample/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config.Load()
	log.Println("Configuration loaded successfully")

	// Initialize database
	if err := models.InitDatabase(models.DatabaseConfig{
		Host:     config.AppConfig.Database.Host,
		User:     config.AppConfig.Database.User,
		Password: config.AppConfig.Database.Password,
		DBName:   config.AppConfig.Database.DBName,
		Port:     config.AppConfig.Database.Port,
		SSLMode:  config.AppConfig.Database.SSLMode,
		TimeZone: config.AppConfig.Database.TimeZone,
	}); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Set Gin mode (can be set via GIN_MODE env var)
	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Add CORS middleware if needed
	// r.Use(cors.Default())

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		auth := v1.Group("/auth")
		{
			// Username/Password authentication
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/refresh", handlers.RefreshAccessToken)
			auth.POST("/logout", handlers.Logout)

			// OAuth2.0 authentication
			oauth := auth.Group("/oauth")
			{
				oauth.GET("/:provider/login", handlers.OAuthLogin)
				oauth.GET("/:provider/callback", handlers.OAuthCallback)
			}
		}

		// Protected routes (authentication required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Current user routes
			protected.GET("/me", handlers.Me)

			// User routes (all authenticated users can view)
			users := protected.Group("/users")
			{
				// Regular users can view user list and specific users
				users.GET("", middleware.RequireRole(models.RoleAdmin, models.RoleUser), handlers.GetUsers)
				users.GET("/:id", middleware.RequireRole(models.RoleAdmin, models.RoleUser), handlers.GetUser)

				// Users can update their own profile (checked in handler)
				users.PUT("/:id", handlers.UpdateUser)

				// Admin-only routes
				users.DELETE("/:id", middleware.RequireAdmin(), handlers.DeleteUser)
				users.PATCH("/:id/role", middleware.RequireAdmin(), handlers.UpdateUserRole)
			}
		}
	}

	// Health check endpoint (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "GoRestExample API",
		})
	})

	// Start server
	port := ":" + config.AppConfig.Server.Port
	log.Printf("Server starting on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
