package meals_handlers

import (
	"math"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/JonathanTriC/nomie-api/internal/database"
	meals_models "github.com/JonathanTriC/nomie-api/internal/modules/meals/models"
	meals_repository "github.com/JonathanTriC/nomie-api/internal/modules/meals/repository"
	meals_services "github.com/JonathanTriC/nomie-api/internal/modules/meals/services"
	"github.com/JonathanTriC/nomie-api/internal/utils"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   meals_services.Service
	repository meals_repository.Repository
	db        *database.Database
	jwtSecret []byte
}

func NewHandler(service meals_services.Service, repository meals_repository.Repository, db *database.Database, jwtSecret []byte) *Handler {
	return &Handler{
		service:   	service,
		repository: repository,
		db:        	db,
		jwtSecret: 	jwtSecret,
	}
}

func (h *Handler) GetTodayRecommendation(c *gin.Context) {
	// Get user ID from context
	userIDInt, ok := utils.GetUserID(c)
	if !ok {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
      return
  }

	userID := strconv.Itoa(userIDInt)

	// Call service
	meal, err := h.service.GetTodayRecommendation(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return JSON
	c.JSON(http.StatusOK, gin.H{
		"meals": []meals_models.Meal{*meal},
	})
}

func (h *Handler) GetPopularPicks(c *gin.Context) {
    userIDInt, ok := utils.GetUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    userID := strconv.Itoa(userIDInt)

    limitStr := c.DefaultQuery("limit", "10")
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 10
    }

    pageStr := c.DefaultQuery("page", "1")
    page, err := strconv.Atoi(pageStr)
    if err != nil || page <= 0 {
        page = 1
    }

    popularPicks, err := h.service.GetPopularPicks(userID, limit, page)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, popularPicks)
}

func (h *Handler) GetCuisinePicks(c *gin.Context) {
    userIDInt, ok := utils.GetUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    userID := strconv.Itoa(userIDInt)

    limitStr := c.DefaultQuery("limit", "10")
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 10
    }

    pageStr := c.DefaultQuery("page", "1")
    page, err := strconv.Atoi(pageStr)
    if err != nil || page <= 0 {
        page = 1
    }

    cuisinePicks, err := h.service.GetCuisinePicks(userID, limit, page)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, cuisinePicks)
}

func (h *Handler) GetSearchMeals(c *gin.Context) {
	// Get user ID from context
	userIDInt, ok := utils.GetUserID(c)
	if !ok {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
      return
  }

	userID := strconv.Itoa(userIDInt)

	query := c.Param("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
			limit = 10
	}

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
			page = 1
	}

	// Call service
	meal, err := h.service.GetSearchMeals(userID, query, limit, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return JSON
	c.JSON(http.StatusOK, meal)
}

func (h *Handler) GetMealDetail(c *gin.Context) {
	userIDInt, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := strconv.Itoa(userIDInt)

	mealID := c.Param("mealId")
	if mealID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mealId is required"})
		return
	}

	meal, err := h.service.GetMealDetail(userID, mealID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go h.repository.LogLastSeen(userID, meal.MealID, meal.MealName, meal.MealThumbImage)

	c.JSON(http.StatusOK, meal)
}

func (h *Handler) GetFavourites(c *gin.Context) {
	userIDInt, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := strconv.Itoa(userIDInt)

	favs, err := h.service.GetFavourites(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"favourites": favs,
	})
}


func (h *Handler) SetFavourite(c *gin.Context) {
	userIDInt, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req meals_models.FavouriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := strconv.Itoa(userIDInt)

	if err := h.service.SetFavourite(userID, req.MealID, req.MealName, req.MealThumbImage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Favourite added"})
}

func (h *Handler) UnsetFavourite(c *gin.Context) {
	userIDInt, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	mealID := c.Param("mealId")
	if mealID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mealId is required"})
		return
	}

	userID := strconv.Itoa(userIDInt)
	if err := h.service.UnsetFavourite(userID, mealID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Favourite removed"})
}

func (h *Handler) GetLastSeen(c *gin.Context) {
    userIDInt, ok := utils.GetUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    userID := strconv.Itoa(userIDInt)

    // limit
    limitStr := c.DefaultQuery("limit", "10")
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 10
    }

    // page
    pageStr := c.DefaultQuery("page", "1")
    page, err := strconv.Atoi(pageStr)
    if err != nil || page <= 0 {
        page = 1
    }

    // call repo with pagination
    meals, totalItems, err := h.repository.GetLastSeen(userID, limit, page)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

    c.JSON(http.StatusOK, gin.H{
        "meals":       meals,
        "page":        page,
        "totalPages":  totalPages,
        "totalItems":  totalItems,
    })
}

func (h *Handler) DeleteLastSeen(c *gin.Context) {
    userIDInt, ok := utils.GetUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    userID := strconv.Itoa(userIDInt)

    err := h.repository.DeleteLastSeen(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Last seen meals deleted successfully"})
}

func (h *Handler) CreateReviewMeals(c *gin.Context) {
    userIDInt, ok := utils.GetUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    userID := strconv.Itoa(userIDInt)

    mealID := c.Param("mealId")
    if mealID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "mealId is required"})
        return
    }

    // parse rating
    ratingStr := c.PostForm("rating")
    rating, err := strconv.Atoi(ratingStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be a number"})
        return
    }
    if rating < 1 || rating > 5 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be between 1 and 5"})
        return
    }

    reviewText := c.PostForm("reviewText")

    // handle optional file upload
    var file multipart.File
    var header *multipart.FileHeader
    file, header, err = c.Request.FormFile("reviewImage")
    if err != nil {
        if err != http.ErrMissingFile {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file upload"})
            return
        }
        file = nil
        header = nil
    }

    review, err := h.service.CreateReview(userID, mealID, rating, reviewText, file, header)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Review created successfully",
        "review":  review,
    })
}

func (h *Handler) GetAllMealsReview(c *gin.Context) {
    mealID := c.Param("mealId")

    reviews, err := h.repository.GetReviewsByMealID(mealID, false) 
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}
