package server

import (
	"github.com/JonathanTriC/nomie-api/internal/config"
	"github.com/JonathanTriC/nomie-api/internal/database"
	"github.com/JonathanTriC/nomie-api/internal/middleware"
	auth_routes "github.com/JonathanTriC/nomie-api/internal/modules/auth/routes"
	meals_routes "github.com/JonathanTriC/nomie-api/internal/modules/meals/routes"
	misc_routes "github.com/JonathanTriC/nomie-api/internal/modules/misc/routes"
	user_routes "github.com/JonathanTriC/nomie-api/internal/modules/user/routes"
	"github.com/gin-gonic/gin"
)

func SetupRouter(db *database.Database, cfg *config.Config) *gin.Engine {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(middleware.CORSMiddleware())

	// Register modules
	auth_routes.RegisterRoute(r, db, cfg.JWT.Secret)
	user_routes.RegisterRoute(r, db, cfg.JWT.Secret)
	meals_routes.RegisterRoutes(r, db, cfg.JWT.Secret)
	misc_routes.RegisterRoutes(r, db, cfg.JWT.Secret)

	return r
}
