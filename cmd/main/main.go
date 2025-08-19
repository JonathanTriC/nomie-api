package main

import (
	"log"
	"net/http"

	"github.com/JonathanTriC/nomie-api/internal/config"
	"github.com/JonathanTriC/nomie-api/internal/database"
	"github.com/JonathanTriC/nomie-api/internal/server"
	"github.com/JonathanTriC/nomie-api/pkg/logger"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Init logger
	logger.Init()
	logger.InfoLogger.Println("Server starting...")

	// Init DB
	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.DB.Close()

	// Init server router
	r := server.SetupRouter(db, cfg)

	// Start server
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Listening on %s", serverAddr)

	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed:", err)
	}
}
