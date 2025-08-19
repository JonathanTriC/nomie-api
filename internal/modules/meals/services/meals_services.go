package meals_services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"time"

	meals_models "github.com/JonathanTriC/nomie-api/internal/modules/meals/models"
	meals_repository "github.com/JonathanTriC/nomie-api/internal/modules/meals/repository"
	misc_services "github.com/JonathanTriC/nomie-api/internal/modules/misc/services"
	"github.com/JonathanTriC/nomie-api/internal/utils"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type Service interface {
	GetTodayRecommendation(userID string) (*meals_models.Meal, error)
	GetPopularPicks(userID string, limit, page int) (map[string]interface{}, error)
	GetCuisinePicks(userID string, limit, page int) (map[string]interface{}, error)
	GetSearchMeals(userID, query string, limit, page int) (map[string]interface{}, error)	
	GetMealDetail(userID, mealID string) (*meals_models.Meal, error)
	GetFavourites(userID string) ([]meals_models.Favourite, error)
	SetFavourite(userID, mealID, mealName, mealThumbImage string) error
	UnsetFavourite(userID, mealID string) error
    CreateReview(userID, mealID string, rating int, reviewText string, reviewImageFile multipart.File, header *multipart.FileHeader) (*meals_models.MealReview, error)
}

type service struct {
	repo meals_repository.Repository
    Cld  *cloudinary.Cloudinary
}

func NewService(repo meals_repository.Repository) Service {
    var apiURL = utils.GetEnv("CDN_API_URL", "your_cdn_api_url")
    cld, err := cloudinary.NewFromURL(apiURL)
    if err != nil {
        log.Fatalf("failed to init cloudinary: %v", err)
    }

    return &service{
        repo:  repo,
        Cld: cld,
    }
}

func (s *service) GetTodayRecommendation(userID string) (*meals_models.Meal, error) {
	resp, err := http.Get("https://www.themealdb.com/api/json/v1/1/random.php")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp meals_models.MealAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if len(apiResp.Meals) == 0 {
		return nil, fmt.Errorf("no meals found")
	}

	return s.buildMealFromAPIData(apiResp.Meals[0], userID)
}

func (s *service) GetPopularPicks(userID string, limit, page int) (map[string]interface{}, error) {
    categories, err := misc_services.NewService().GetCategories()
    if err != nil {
        return nil, err
    }
    if len(categories) == 0 {
        return nil, fmt.Errorf("no categories found")
    }

    dayIndex := time.Now().YearDay() % len(categories)
    category := categories[dayIndex]

    url := fmt.Sprintf("https://www.themealdb.com/api/json/v1/1/filter.php?c=%s", category)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var mealsResp meals_models.MealAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&mealsResp); err != nil {
        return nil, err
    }

    totalItems := len(mealsResp.Meals)
    totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

    // Pagination math
    start := (page - 1) * limit
    end := start + limit
    if start > totalItems {
        return map[string]interface{}{
            "categoryName": category,
            "meals":        []map[string]interface{}{},
            "page":         page,
            "totalItems":   totalItems,
            "totalPages":   totalPages,
        }, nil
    }
    if end > totalItems {
        end = totalItems
    }

    var meals []map[string]interface{}
    for _, m := range mealsResp.Meals[start:end] {
        idMeal, _ := m["idMeal"].(string)
        isFav, _ := s.repo.IsFavourite(userID, idMeal)

        meals = append(meals, map[string]interface{}{
            "mealName":       m["strMeal"],
            "mealThumbImage": m["strMealThumb"],
            "mealId":         idMeal,
            "isFavourite":    isFav,
        })
    }

    return map[string]interface{}{
        "categoryName": category,
        "meals":        meals,
        "page":         page,
        "totalItems":   totalItems,
        "totalPages":   totalPages,
    }, nil
}

func (s *service) GetCuisinePicks(userID string, limit, page int) (map[string]interface{}, error) {
    areas, err := misc_services.NewService().GetAreas()
    if err != nil {
        return nil, err
    }
    if len(areas) == 0 {
        return nil, fmt.Errorf("no areas found")
    }

    weekIndex := func() int {
				_, week := time.Now().ISOWeek() 
				return week % len(areas)        
		}()
		area := areas[weekIndex]

    url := fmt.Sprintf("https://www.themealdb.com/api/json/v1/1/filter.php?a=%s", area)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var mealsResp meals_models.MealAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&mealsResp); err != nil {
        return nil, err
    }

    totalItems := len(mealsResp.Meals)
    totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

    // Pagination math
    start := (page - 1) * limit
    end := start + limit
    if start > totalItems {
        return map[string]interface{}{
            "areasName": area,
            "meals":        []map[string]interface{}{},
            "page":         page,
            "totalItems":   totalItems,
            "totalPages":   totalPages,
        }, nil
    }
    if end > totalItems {
        end = totalItems
    }

    var meals []map[string]interface{}
    for _, m := range mealsResp.Meals[start:end] {
        idMeal, _ := m["idMeal"].(string)
        isFav, _ := s.repo.IsFavourite(userID, idMeal)

        meals = append(meals, map[string]interface{}{
            "mealName":       m["strMeal"],
            "mealThumbImage": m["strMealThumb"],
            "mealId":         idMeal,
            "isFavourite":    isFav,
        })
    }

    return map[string]interface{}{
        "areaName": area,
        "meals":        meals,
        "page":         page,
        "totalItems":   totalItems,
        "totalPages":   totalPages,
    }, nil
}

func (s *service) GetSearchMeals(userID, query string, limit, page int) (map[string]interface{}, error) {
    url := fmt.Sprintf("https://www.themealdb.com/api/json/v1/1/search.php?s=%s", query)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var apiResp meals_models.MealAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, err
    }
    if len(apiResp.Meals) == 0 {
        return nil, fmt.Errorf("no meals found")
    }

    totalItems := len(apiResp.Meals)
    totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

    start := (page - 1) * limit
    end := start + limit
    if start > totalItems {
        return map[string]interface{}{
            "meals":      []map[string]interface{}{},
            "page":       page,
            "totalItems": totalItems,
            "totalPages": totalPages,
        }, nil
    }
    if end > totalItems {
        end = totalItems
    }

    var meals []map[string]interface{}
    for _, meal := range apiResp.Meals[start:end] {
        builtMeal, err := s.buildMealFromAPIData(meal, userID)
        if err != nil {
            return nil, err
        }

        meals = append(meals, map[string]interface{}{
            "mealId":         builtMeal.MealID,
            "mealName":       builtMeal.MealName,
            "mealThumbImage": builtMeal.MealThumbImage,
            "isFavourite":    builtMeal.IsFavourite,
        })
    }

    return map[string]interface{}{
        "meals":      meals,
        "page":       page,
        "totalItems": totalItems,
        "totalPages": totalPages,
    }, nil
}


func (s *service) GetMealDetail(userID, mealID string) (*meals_models.Meal, error) {
	url := fmt.Sprintf("https://www.themealdb.com/api/json/v1/1/lookup.php?i=%s", mealID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp meals_models.MealAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if len(apiResp.Meals) == 0 {
		return nil, fmt.Errorf("no meals found for id %s", mealID)
	}

	meal, err := s.buildMealFromAPIData(apiResp.Meals[0], userID)
	if err != nil {
		return nil, err
	}

	// get reviews from repo
	reviews, err := s.repo.GetReviewsByMealID(mealID, true)
	if err != nil {
		return nil, err
	}
	meal.Reviews = reviews

    avgRating, totalReviews, err := s.repo.GetMealRating(mealID)
    if err != nil {
        return nil, err
    }
    meal.AvgRating = avgRating
    meal.TotalReviews = totalReviews

	return meal, nil
}


func (s *service) GetFavourites(userID string) ([]meals_models.Favourite, error) {
	return s.repo.GetFavourites(userID)
}

func (s *service) SetFavourite(userID, mealID, mealName, mealThumbImage string) error {
	return s.repo.AddFavourite(userID, mealID, mealName, mealThumbImage)
}

func (s *service) UnsetFavourite(userID, mealID string) error {
	return s.repo.RemoveFavourite(userID, mealID)
}

func (s *service) CreateReview(userID, mealID string, rating int, reviewText string, reviewImageFile multipart.File, header *multipart.FileHeader) (*meals_models.MealReview, error) {
	var imageURL string

	// upload to CDN if file provided
	if reviewImageFile != nil {
		ctx := context.Background()
		uploadResult, err := s.Cld.Upload.Upload(ctx, reviewImageFile, uploader.UploadParams{
			Folder:   "reviews",
			PublicID: fmt.Sprintf("review_%s_%s", mealID, userID),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upload image: %w", err)
		}
		imageURL = uploadResult.SecureURL
	}

	review := &meals_models.MealReview{
		UserID:      userID,
		MealID:      mealID,
		Rating:      rating,
		ReviewText:  reviewText,
		ReviewImage: imageURL,
	}

	if err := s.repo.CreateReview(review); err != nil {
		return nil, err
	}

	return review, nil
}

func valOrEmpty(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func (s *service) buildMealFromAPIData(m map[string]interface{}, userID string) (*meals_models.Meal, error) {
	meal := &meals_models.Meal{
		MealID:                       valOrEmpty(m["idMeal"]),
		MealName:                     valOrEmpty(m["strMeal"]),
		MealAlternate:                valOrEmpty(m["strMealAlternate"]),
		MealCategory:                 valOrEmpty(m["strCategory"]),
		MealArea:                     valOrEmpty(m["strArea"]),
		MealInstructions:             valOrEmpty(m["strInstructions"]),
		MealThumbImage:               valOrEmpty(m["strMealThumb"]),
		MealTags:                     valOrEmpty(m["strTags"]),
		MealYoutubeTutorial:          valOrEmpty(m["strYoutube"]),
		MealSource:                   valOrEmpty(m["strSource"]),
		MealImageSource:              valOrEmpty(m["strImageSource"]),
		MealCreativeCommonsConfirmed: valOrEmpty(m["strCreativeCommonsConfirmed"]),
		DateModified:                 valOrEmpty(m["dateModified"]),
	}

	// Ingredients mapping
	for i := 1; i <= 20; i++ {
		ingKey := fmt.Sprintf("strIngredient%d", i)
		meaKey := fmt.Sprintf("strMeasure%d", i)
		ingredient := valOrEmpty(m[ingKey])
		measure := valOrEmpty(m[meaKey])
		if ingredient != "" {
			meal.MealIngredient = append(meal.MealIngredient, meals_models.MealIngredient{
				IngredientName:    ingredient,
				IngredientMeasure: measure,
			})
		}
	}

	// Check favourite
	isFav, err := s.repo.IsFavourite(userID, meal.MealID)
	if err != nil {
		return nil, err
	}
	meal.IsFavourite = isFav

	return meal, nil
}
