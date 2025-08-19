package meals_routes

import (
	"github.com/gin-gonic/gin"

	"github.com/JonathanTriC/nomie-api/internal/database"
	auth_middleware "github.com/JonathanTriC/nomie-api/internal/modules/auth/middleware"
	meals_handlers "github.com/JonathanTriC/nomie-api/internal/modules/meals/handlers"
	meals_repository "github.com/JonathanTriC/nomie-api/internal/modules/meals/repository"
	meals_services "github.com/JonathanTriC/nomie-api/internal/modules/meals/services"
)

func RegisterRoutes(r *gin.Engine, db *database.Database, jwtSecret string) {
	// Initialize repository
	repo := meals_repository.NewRepository(db)

	service := meals_services.NewService(repo)

	// Initialize handler
	handler := meals_handlers.NewHandler(service, repo, db, []byte(jwtSecret))

	// Define routes
	mealGroup := r.Group("/v1/meals")
	{
		protected := mealGroup.Group("")
		protected.Use(auth_middleware.AuthMiddleware([]byte(jwtSecret)))
		{
			// MARK: Chef’s Pick of the Day
			protected.GET("/today-recommendation", handler.GetTodayRecommendation)
			// MARK: Tasty {Category} Dishes Everyone Loves
			protected.GET("/popular-picks", handler.GetPopularPicks)
			// MARK: Straight from {Area} Kitchens
			protected.GET("/cuisine-picks", handler.GetCuisinePicks)
			protected.GET("/search/:query", handler.GetSearchMeals)
			protected.GET("/detail/:mealId", handler.GetMealDetail)
			// MARK: Your Tasty Collection
			protected.GET("/favourites", handler.GetFavourites)
			protected.POST("/favourites", handler.SetFavourite)
			protected.DELETE("/favourites/:mealId", handler.UnsetFavourite)
			// MARK: Back for Another Bite?
			protected.GET("/last-seen", handler.GetLastSeen)
			protected.DELETE("/last-seen", handler.DeleteLastSeen)

			protected.POST("/reviews/:mealId", handler.CreateReviewMeals)
			protected.GET("/all-reviews/:mealId", handler.GetAllMealsReview)
		}
	}
}
