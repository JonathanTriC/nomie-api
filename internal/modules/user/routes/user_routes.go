package user_routes

import (
	"github.com/JonathanTriC/nomie-api/internal/database"
	auth_middleware "github.com/JonathanTriC/nomie-api/internal/modules/auth/middleware"
	user_handlers "github.com/JonathanTriC/nomie-api/internal/modules/user/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoute(r *gin.Engine, db *database.Database, jwtSecret string) {
	handler := user_handlers.NewUserHandler(db, []byte(jwtSecret))

	userGroup := r.Group("/v1/user")
	{
		protected := userGroup.Group("")
		protected.Use(auth_middleware.AuthMiddleware([]byte(jwtSecret)))
		{
			protected.GET("/profile", handler.GetUserProfile)
			protected.POST("/update-profile", handler.UpdateProfile)
			protected.POST("/change-password", handler.ChangePassword)
			protected.POST("/delete-account", handler.DeleteAccount)
		}
	}
}
