package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
    Server struct {
        Port         string
        Host         string
        ReadTimeout  time.Duration
        WriteTimeout time.Duration
    }
    
    Database struct {
        Host     string
        Port     string
        User     string
        Password string
        DBName   string
        SSLMode  string
    }
    
    JWT struct {
        Secret        string
        TokenExpiry   time.Duration
        RefreshExpiry time.Duration
    }
    
    Environment string
}

func Load() (*Config, error) {
    godotenv.Load() // Load .env if exists
    
    cfg := &Config{}
    
    // Server config
    cfg.Server.Port = getEnv("SERVER_PORT", "your_server_port")
    cfg.Server.Host = getEnv("SERVER_HOST", "your_server_host")
    cfg.Server.ReadTimeout = time.Second * 15
    cfg.Server.WriteTimeout = time.Second * 15
    
    // Database config
    cfg.Database.Host = getEnv("DB_HOST", "your_database_host")
    cfg.Database.Port = getEnv("DB_PORT", "your_database_port")
    cfg.Database.User = getEnv("DB_USER", "your_database_user")
    cfg.Database.Password = getEnv("DB_PASSWORD", "your_database_password")
    cfg.Database.DBName = getEnv("DB_NAME", "your_database_name")
    cfg.Database.SSLMode = getEnv("DB_SSLMODE", "your_database_ssl_mode")
    
    // JWT config
    cfg.JWT.Secret = getEnv("JWT_SECRET", "your_jwt_secret")
    cfg.JWT.TokenExpiry = time.Hour * 24    // 24 hours
    cfg.JWT.RefreshExpiry = time.Hour * 168 // 7 days
    
    cfg.Environment = getEnv("ENV", "development")
    
    return cfg, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func (c *Config) GetDSN() string {
    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        c.Database.Host,
        c.Database.Port,
        c.Database.User,
        c.Database.Password,
        c.Database.DBName,
        c.Database.SSLMode,
    )
}