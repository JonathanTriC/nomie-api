package auth_handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/JonathanTriC/nomie-api/internal/database"
	auth_models "github.com/JonathanTriC/nomie-api/internal/modules/auth/models"
	"github.com/JonathanTriC/nomie-api/internal/utils"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var tokenBlacklist = make(map[string]bool)

type Claims struct {
	UserID      int    `json:"user_id"`
	Email       string `json:"email"`
	jwt.RegisteredClaims
}


type AuthHandler struct {
    db              *database.Database
    jwtSecret       []byte
    tokenExpiration time.Duration
}

func NewAuthHandler(db *database.Database, jwtSecret []byte) *AuthHandler {
    return &AuthHandler{
        db:              db,
        jwtSecret:       jwtSecret,
        tokenExpiration: 24 * time.Hour,
    }
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
    godotenv.Load()

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // max memory buffer
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	// Get avatar
	file, header, err := c.Request.FormFile("avatar")
	if err != nil && err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file upload"})
		return
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	var avatarURL string
	if file != nil {
		// Validate file size (max 5MB)
		if header.Size > 5*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File size must be less than 5MB"})
			return
		}

		// Validate file extension
		ext := filepath.Ext(header.Filename)
		allowedExt := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
		}
		if !allowedExt[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File must be .jpg, .jpeg, or .png"})
			return
		}

		// Upload to Cloudinary
        var cloudName = utils.GetEnv("CDN_CLOUD_NAME", "your_cdn_cloud_name")
        var apiKey = utils.GetEnv("CDN_API_KEY", "your_cdn_api_key")
        var apiSecret = utils.GetEnv("CDN_API_SECRET", "your_cdn_api_secret")
		cld, _ := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
		uploadResult, err := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{
			Folder: "avatars",
			PublicID: fmt.Sprintf("avatar_%s", utils.GenerateRandomID()),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload avatar"})
			return
		}
		avatarURL = uploadResult.SecureURL
	}

	// Bind other fields
	user := auth_models.UserRegister{
		Username: c.PostForm("username"),
		Fullname: c.PostForm("fullname"),
		Email:    c.PostForm("email"),
		Password: c.PostForm("password"),
	}

	if file == nil {
	parts := strings.Fields(user.Fullname)

	firstname := ""
	lastname := ""

	if len(parts) > 0 {
		firstname = parts[0]
	}
	if len(parts) > 1 {
		lastname = parts[len(parts)-1]
	}

	// assign instead of redeclare
	avatarURL = "https://avatar.iran.liara.run/username?username=" + firstname
	if lastname != "" {
		avatarURL += "+" + lastname
	}

	user.Avatar = avatarURL
} else {
	user.Avatar = avatarURL
}

	if err := user.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatErrorRegister(err)})
		return
	}

	// Check email exists
	var exists bool
	err = h.db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, user.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Check username exists
	err = h.db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, user.Username).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password processing failed"})
		return
	}

	// Insert user
	var id int
	err = h.db.DB.QueryRow(`
        INSERT INTO users (username, fullname, email, avatar, password_hash) 
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`,
		user.Username, user.Fullname, user.Email, user.Avatar, hashedPassword,
	).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User creation failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": id,
		"avatar":  avatarURL,
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
    var login auth_models.UserLogin
    if err := c.ShouldBindJSON(&login); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login data"})
        return
    }

    var user auth_models.User
    err := h.db.DB.QueryRow(`
        SELECT id, username, fullname, email, avatar, password_hash 
        FROM users 
        WHERE email = $1`,
        login.Email,
    ).Scan(&user.ID, &user.Username, &user.Fullname, &user.Email, &user.Avatar, &user.PasswordHash)

    if err == sql.ErrNoRows {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Login process failed"})
        return
    }

    if !utils.CheckPasswordHash(login.Password, user.PasswordHash) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    now := time.Now()
    claims := jwt.MapClaims{
        "user_id":  user.ID,
        "username": user.Username,
        "fullname": user.Fullname,
        "email":    user.Email,
        "avatar":   user.Avatar,
        "iat":      now.Unix(),
        "exp":      now.Add(h.tokenExpiration).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(h.jwtSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "token":      tokenString,
        "expires_in": h.tokenExpiration.Seconds(),
        "token_type": "Bearer",
    })
}

// CheckEmail verifies whether an email is already registered
func (h *AuthHandler) CheckEmail(c *gin.Context) {
    type request struct {
        Email string `json:"email" binding:"required,email"`
    }
    var req request

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
        return
    }

    var exists bool
    err := h.db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, req.Email).Scan(&exists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "isRegistered": exists,
    })
}


// RefreshToken issues a new access token using a valid refresh token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	// Parse and validate refresh token
	token, err := jwt.ParseWithClaims(request.RefreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.ExpiresAt.Time.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Expired or invalid refresh token"})
		return
	}

	// Generate a new access token
	newAccessToken, err := h.generateAccessToken(claims.UserID, claims.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
	})
}

// Logout invalidates the current refresh token
func (h *AuthHandler) Logout(c *gin.Context) {
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
        return
    }

    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    if tokenString == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Authorization header"})
        return
    }

    tokenBlacklist[tokenString] = true

    c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}


// generateAccessToken creates a short-lived JWT access token
func (h *AuthHandler) generateAccessToken(userID int, email string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // access token valid for 15 minutes
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}