package auth_routes

import (
	"github.com/JonathanTriC/nomie-api/internal/database"
	auth_handlers "github.com/JonathanTriC/nomie-api/internal/modules/auth/handlers"
	auth_middleware "github.com/JonathanTriC/nomie-api/internal/modules/auth/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoute(r *gin.Engine, db *database.Database, jwtSecret string) {
	handler := auth_handlers.NewAuthHandler(db, []byte(jwtSecret))

	authGroup := r.Group("/v1/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/check-email", handler.CheckEmail)

		protected := authGroup.Group("")
		protected.Use(auth_middleware.AuthMiddleware([]byte(jwtSecret)))
		{
			protected.POST("/refresh-token", handler.RefreshToken)
			protected.POST("/logout", handler.Logout)
		}
	}
}
