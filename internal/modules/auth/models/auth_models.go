package auth_models

import (
	"errors"
	"regexp"
	"time"
)

// User represents a database user
type User struct {
    ID           int       `json:"id"`
    Email        string    `json:"email"`
    Fullname     string    `json:"fullname"`
    Username     string    `json:"username"`
    Avatar       string    `json:"avatar"`
    PasswordHash string    `json:"-"` // exclude from JSON
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// UserLogin represents login request data
type UserLogin struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// UserRegister represents registration request data
type UserRegister struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Fullname string `json:"fullname" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Avatar   string `json:"avatar" binding:"omitempty,url"` // optional
}

// Validate checks if email format is valid (extra manual validation)
func (u *UserRegister) Validate() error {
    emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
    if !emailRegex.MatchString(u.Email) {
        return errors.New("invalid email format")
    }
    return nil
}
