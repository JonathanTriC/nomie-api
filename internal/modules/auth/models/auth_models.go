package auth_models

import (
	"regexp"
	"strings"
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

type ValidationError struct {
    FormKey string `json:"form_key"`
    Message string `json:"error"`
}

func (e *ValidationError) Error() string {
    return e.Message
}

// Validate checks if email format is valid (extra manual validation)
func (u *UserRegister) Validate() error {
    // Email validation
    emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
    if !emailRegex.MatchString(strings.ToLower(u.Email)) {
        return &ValidationError{FormKey: "email", Message: "Please enter a valid email address"}
    }

    // Username validation
    if len(u.Username) < 3 || len(u.Username) > 20 {
        return &ValidationError{FormKey: "username", Message: "Your username needs to be between 3 and 20 characters long."}
    }
    usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
    if !usernameRegex.MatchString(u.Username) {
        return &ValidationError{FormKey: "username", Message: "Usernames can only contain letters, numbers, and underscores (_)."}
    }

    // Fullname validation
    if len(u.Fullname) == 0 || len(u.Fullname) > 30 {
        return &ValidationError{FormKey: "fullname", Message: "Please enter your name (up to 30 characters)."}
    }
    fullnameRegex := regexp.MustCompile(`^[a-zA-Z ]+$`)
    if !fullnameRegex.MatchString(u.Fullname) {
        return &ValidationError{FormKey: "fullname", Message: "Your name can only include letters and spaces."}
    }

    // Password validation
    if len(u.Password) < 8 {
        return &ValidationError{FormKey: "password", Message: "Your password must be at least 8 characters long to be secure."}
    }
    passwordUpper := regexp.MustCompile(`[A-Z]`)
    passwordLower := regexp.MustCompile(`[a-z]`)
    passwordDigit := regexp.MustCompile(`[0-9]`)
    passwordSpecial := regexp.MustCompile(`[!@#\$%\^&\*]`)

    if !passwordUpper.MatchString(u.Password) ||
        !passwordLower.MatchString(u.Password) ||
        !passwordDigit.MatchString(u.Password) ||
        !passwordSpecial.MatchString(u.Password) {
        return &ValidationError{
            FormKey: "password",
            Message: "For a stronger password, please include uppercase, lowercase, a number, and a symbol",
        }
    }

    return nil
}

