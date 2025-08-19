package user_handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/JonathanTriC/nomie-api/internal/database"
	"github.com/JonathanTriC/nomie-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type UserHandler struct {
    db              *database.Database
    jwtSecret       []byte
    tokenExpiration time.Duration
}

func NewUserHandler(db *database.Database, jwtSecret []byte) *UserHandler {
    return &UserHandler{
        db:              db,
        jwtSecret:       jwtSecret,
        tokenExpiration: 24 * time.Hour,
    }
}

// UpdateProfile allows authenticated users to update their profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    var req struct {
        Fullname string `json:"fullname"`
        Username string `json:"username"`
        Avatar   string `json:"avatar"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    _, err := h.db.DB.Exec(`
        UPDATE users 
        SET fullname = $1, username = $2, avatar = $3 
        WHERE id = $4
    `, req.Fullname, req.Username, req.Avatar, userID)
    if err != nil {
        // Check for Postgres unique constraint violation
        if pqErr, ok := err.(*pq.Error); ok {
            if pqErr.Code == "23505" { // unique_violation
                if strings.Contains(pqErr.Constraint, "users_username_key") {
                    c.JSON(http.StatusConflict, gin.H{"error": "Username is already taken"})
                    return
                }
            }
        }

        // Fallback for any other DB error
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, please try again"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}


// ChangePassword allows authenticated users to change their password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

	var input struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatErrorRegister(err)})
		return
	}

	var storedHash string
	err := h.db.DB.QueryRow(`SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&storedHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !utils.CheckPasswordHash(input.CurrentPassword, storedHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	newHash, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password processing failed"})
		return
	}

	_, err = h.db.DB.Exec(`UPDATE users SET password_hash = $1 WHERE id = $2`, newHash, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// DeleteAccount allows authenticated users to permanently delete their account
func (h *UserHandler) DeleteAccount(c *gin.Context) {
    userID, ok := utils.GetUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    _, err := h.db.DB.Exec(`DELETE FROM users WHERE id = $1`, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
    userID, _ := c.Get("user_id")
    email, _ := c.Get("email")
    username, _ := c.Get("username")
    fullname, _ := c.Get("fullname")
    avatar, _ := c.Get("avatar")
    
    c.JSON(200, gin.H{
        "user_id": userID,
        "email":   email,
        "username":   username,
        "fullname":   fullname,
        "avatar":   avatar,
    })
}
