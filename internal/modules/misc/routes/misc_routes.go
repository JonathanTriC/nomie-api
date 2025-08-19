package misc_routes

import (
	"github.com/gin-gonic/gin"

	"github.com/JonathanTriC/nomie-api/internal/database"
	auth_middleware "github.com/JonathanTriC/nomie-api/internal/modules/auth/middleware"
	misc_handlers "github.com/JonathanTriC/nomie-api/internal/modules/misc/handlers"
	misc_services "github.com/JonathanTriC/nomie-api/internal/modules/misc/services"
)

func RegisterRoutes(r *gin.Engine, db *database.Database, jwtSecret string) {
	service := misc_services.NewService()

	// Initialize handler
	handler := misc_handlers.NewHandler(service)

	// Define routes
	mealGroup := r.Group("/v1/misc")
	{
		protected := mealGroup.Group("")
		protected.Use(auth_middleware.AuthMiddleware([]byte(jwtSecret)))
		{
			protected.GET("/category", handler.GetCategoryList)
			protected.GET("/area", handler.GetAreaList)
		}
	}
}
