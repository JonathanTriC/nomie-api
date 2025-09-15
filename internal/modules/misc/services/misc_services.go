package misc_services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	misc_models "github.com/JonathanTriC/nomie-api/internal/modules/misc/models"
)

type Service interface {
	GetCategories() ([]string, error)
	GetAreas() ([]Area, error)
}
type Area struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	ImageURL string `json:"imageUrl"`
}

type apiMealResponse struct {
	Meals []map[string]string `json:"meals"`
}


type service struct{} 

var areaToCodeMap = map[string]string{
	"American":   "US",
	"British":    "GB",
	"Canadian":   "CA",
	"Chinese":    "CN",
	"Croatian":   "HR",
	"Dutch":      "NL",
	"Egyptian":   "EG",
	"Filipino":   "PH",
	"French":     "FR",
	"Greek":      "GR",
	"Indian":     "IN",
	"Irish":      "IE",
	"Italian":    "IT",
	"Jamaican":   "JM",
	"Japanese":   "JP",
	"Kenyan":     "KE",
	"Malaysian":  "MY",
	"Mexican":    "MX",
	"Moroccan":   "MA",
	"Polish":     "PL",
	"Portuguese": "PT",
	"Russian":    "RU",
	"Spanish":    "ES",
	"Thai":       "TH",
	"Tunisian":   "TN",
	"Turkish":    "TR",
	"Ukrainian":  "UA",
	"Uruguayan":  "UY",
	"Vietnamese": "VN",
}


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

func (s *service) GetAreas() ([]Area, error) {
	resp, err := http.Get("https://www.themealdb.com/api/json/v1/1/list.php?a=list")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp apiMealResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	var areas []Area
	for _, m := range apiResp.Meals {
		areaName := m["strArea"]
		code := areaToCodeMap[areaName]
		
		var imageUrl string
		if code != "" {
			imageUrl = fmt.Sprintf("https://www.themealdb.com/images/icons/flags/big/64/%s.png", strings.ToLower(code))
		}
		
		areas = append(areas, Area{
			Name:     areaName,
			Code:     code,
			ImageURL: imageUrl,
		})
	}

	return areas, nil
}

