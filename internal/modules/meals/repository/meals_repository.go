package meals_repository

import (
	"math"

	"github.com/JonathanTriC/nomie-api/internal/database"
	meals_models "github.com/JonathanTriC/nomie-api/internal/modules/meals/models"
)

type Repository interface {
	IsFavourite(userID, mealID string) (bool, error)
	GetFavourites(userID string) ([]meals_models.Favourite, error)
	AddFavourite(userID string, mealID, mealName, mealThumbImage string) error
	RemoveFavourite(userID, mealID string) error
	LogLastSeen(userID, mealID, mealName, mealThumb string) error
	GetLastSeen(userID string, limit, page int) ([]map[string]interface{}, int, error)
	DeleteLastSeen(userID string) error
    CreateReview(review *meals_models.MealReview) error
    GetReviewsByMealID(mealID string, limitReviews bool) ([]meals_models.MealReview, error)
    GetMealRating(mealID string) (float64, int, error)
}

type repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) Repository {
	return &repository{db: db}
}

func (r *repository) IsFavourite(userID, mealID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM favourites WHERE user_id=$1 AND meal_id=$2)`
	err := r.db.QueryRow(query, userID, mealID).Scan(&exists)
	return exists, err
}

func (r *repository) GetFavourites(userID string) ([]meals_models.Favourite, error) {
	rows, err := r.db.Query(`
		SELECT meal_id, meal_name, meal_thumb
		FROM favourites
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favourites []meals_models.Favourite
	for rows.Next() {
		var f meals_models.Favourite
		if err := rows.Scan(&f.MealID, &f.MealName, &f.MealThumbImage); err != nil {
			return nil, err
		}
		favourites = append(favourites, f)
	}

	return favourites, nil
}


func (r *repository) AddFavourite(userID string, mealID, mealName, mealThumbImage string) error {
	_, err := r.db.Exec(`
		INSERT INTO favourites (user_id, meal_id, meal_name, meal_thumb)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, meal_id) DO NOTHING
	`, userID, mealID, mealName, mealThumbImage)
	return err
}

func (r *repository) RemoveFavourite(userID, mealID string) error {
	_, err := r.db.Exec(`
		DELETE FROM favourites
		WHERE user_id = $1 AND meal_id = $2
	`, userID, mealID)
	return err
}

func (r *repository) LogLastSeen(userID, mealID, mealName, mealThumbImage string) error {
    query := `
        INSERT INTO user_last_seen (user_id, meal_id, meal_name, meal_thumb_image)
        VALUES ($1, $2, $3, $4)
    `
    _, err := r.db.Exec(query, userID, mealID, mealName, mealThumbImage)
    return err
}

func (r *repository) GetLastSeen(userID string, limit, page int) ([]map[string]interface{}, int, error) {
    // count total
    var totalItems int
    err := r.db.QueryRow(`
        SELECT COUNT(*) 
        FROM user_last_seen 
        WHERE user_id = $1
    `, userID).Scan(&totalItems)
    if err != nil {
        return nil, 0, err
    }

    // pagination math
    offset := (page - 1) * limit

    // fetch meals
    rows, err := r.db.Query(`
        SELECT meal_id, meal_name, meal_thumb_image
        FROM user_last_seen
        WHERE user_id = $1
        ORDER BY seen_at DESC
        LIMIT $2 OFFSET $3
    `, userID, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var meals []map[string]interface{}
    for rows.Next() {
        var mealID, mealName, mealThumb string
        if err := rows.Scan(&mealID, &mealName, &mealThumb); err != nil {
            return nil, 0, err
        }

        isFav, _ := r.IsFavourite(userID, mealID)

        meals = append(meals, map[string]interface{}{
            "mealId":         mealID,
            "mealName":       mealName,
            "mealThumbImage": mealThumb,
            "isFavourite":    isFav,
        })
    }

    return meals, totalItems, nil
}

func (r *repository) DeleteLastSeen(userID string) error {
    query := `DELETE FROM user_last_seen WHERE user_id = $1`
    _, err := r.db.Exec(query, userID)
    return err
}

func (r *repository) CreateReview(review *meals_models.MealReview) error {
	_, err := r.db.Exec(`
		INSERT INTO meal_reviews (user_id, meal_id, rating, review_text, review_image, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, review.UserID, review.MealID, review.Rating, review.ReviewText, review.ReviewImage)
	return err
}

func (r *repository) GetReviewsByMealID(mealID string, limitReviews bool) ([]meals_models.MealReview, error) {
    var reviewQuery string
    if limitReviews {
        reviewQuery = `
            SELECT r.id, r.user_id, r.meal_id, r.rating, r.review_text, r.review_image, r.created_at
            FROM meal_reviews r
            WHERE r.meal_id = $1
            ORDER BY r.created_at DESC
            LIMIT 3
        `
    } else {
        reviewQuery = `
            SELECT r.id, r.user_id, r.meal_id, r.rating, r.review_text, r.review_image, r.created_at
            FROM meal_reviews r
            WHERE r.meal_id = $1
            ORDER BY r.created_at DESC
        `
    }

	rows, err := r.db.Query(reviewQuery, mealID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []meals_models.MealReview
	for rows.Next() {
		var r meals_models.MealReview
		if err := rows.Scan(
            &r.ID,
            &r.UserID,
            &r.MealID,
            &r.Rating,
            &r.ReviewText,
            &r.ReviewImage,
            &r.CreatedAt,
        ); err != nil {
            return nil, err
        }
		reviews = append(reviews, r)
	}
	return reviews, nil
}

func (r *repository) GetMealRating(mealID string) (float64, int, error) {
    var avgRating float64
    var totalReviews int

    err := r.db.QueryRow(`
        SELECT COALESCE(AVG(rating), 0), COUNT(*)
        FROM meal_reviews
        WHERE meal_id = $1
    `, mealID).Scan(&avgRating, &totalReviews)

    if err != nil {
        return 0, 0, err
    }

    avgRating = math.Round(avgRating*10) / 10

    return avgRating, totalReviews, nil
}
