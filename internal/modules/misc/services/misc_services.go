package misc_services

import (
	"encoding/json"
	"net/http"

	misc_models "github.com/JonathanTriC/nomie-api/internal/modules/misc/models"
)

type Service interface {
	GetCategories() ([]string, error)
	GetAreas() ([]string, error)
}

type service struct{} 

func NewService() Service {
	return &service{}
}

func (s *service) GetCategories() ([]string, error) {
	resp, err := http.Get("https://www.themealdb.com/api/json/v1/1/list.php?c=list")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp misc_models.CategoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	var categories []string
	for _, m := range apiResp.Meals {
		categories = append(categories, m["strCategory"])
	}

	return categories, nil
}

func (s *service) GetAreas() ([]string, error) {
	resp, err := http.Get("https://www.themealdb.com/api/json/v1/1/list.php?a=list")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp misc_models.CategoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	var areas []string
	for _, m := range apiResp.Meals {
		areas = append(areas, m["strArea"])
	}

	return areas, nil
}
