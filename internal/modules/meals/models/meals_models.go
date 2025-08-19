package meals_models

import "time"

type Meal struct {
	MealID                     string           `json:"mealId"`
	MealName                   string           `json:"mealName"`
	MealAlternate              string           `json:"mealAlternate"`
	MealCategory               string           `json:"mealCategory"`
	MealArea                   string           `json:"mealArea"`
	MealInstructions           string           `json:"mealInstructions"`
	MealThumbImage             string           `json:"mealThumbImage"`
	MealTags                   string           `json:"mealTags"`
	MealYoutubeTutorial        string           `json:"mealYoutubeTutorial"`
	MealIngredient             []MealIngredient `json:"mealIngredient"`
	MealSource                 string           `json:"mealSource"`
	MealImageSource            string           `json:"mealImageSource"`
	MealCreativeCommonsConfirmed string         `json:"mealCreativeCommonsConfirmed"`
	DateModified               string           `json:"dateModified"`
	IsFavourite                bool             `json:"isFavourite"`
	Reviews 									 []MealReview 		`json:"reviews,omitempty"`
	AvgRating     						 float64          `json:"avgRating"`
  TotalReviews  						 int              `json:"totalReviews"`

}

type MealIngredient struct {
	IngredientName   string `json:"ingredientName"`
	IngredientMeasure string `json:"ingredientMeasure"`
}

// Structs for mapping API response
type MealAPIResponse struct {
	Meals []map[string]interface{} `json:"meals"`
}

type Favourite struct {
	MealID         string `json:"mealId"`
	MealName       string `json:"mealName"`
	MealThumbImage string `json:"mealThumbImage"`
}

type FavouriteRequest struct {
	MealID    string `json:"mealId" binding:"required"`
	MealName  string `json:"mealName" binding:"required"`
	MealThumbImage string `json:"mealThumbImage" binding:"required"`
}

type MealReview struct {
	ID          int       `json:"id" db:"id"`
	UserID      string    `json:"userId" db:"user_id"`
	MealID      string    `json:"mealId" db:"meal_id"`
	Rating      int       `json:"rating" db:"rating"`
	ReviewText  string    `json:"reviewText" db:"review_text"`
	ReviewImage string    `json:"reviewImage,omitempty" db:"review_image"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}
