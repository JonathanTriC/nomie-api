package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func FormatErrorRegister(err error) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, e := range errs {
			switch e.Tag() {
			case "required":
				messages = append(messages, fmt.Sprintf("%s is required", e.Field()))
			case "email":
				messages = append(messages, "Invalid email format")
			case "min":
				messages = append(messages, fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param()))
			default:
				messages = append(messages, fmt.Sprintf("%s is invalid", e.Field()))
			}
		}
		return strings.Join(messages, ", ")
	}
	return err.Error()
}

func GetUserID(c *gin.Context) (int, bool) {
    idVal, exists := c.Get("user_id")
    if !exists {
        return 0, false
    }
    switch v := idVal.(type) {
    case int:
        return v, true
    case int64:
        return int(v), true
    case float64:
        return int(v), true
    default:
        return 0, false
    }
}

func GetEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func GenerateRandomID() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    return hex.EncodeToString(b)
}