package main

import (
	"car_service/config"
	"car_service/database"
	"car_service/logger"
	"car_service/server"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	// Initialize logger with configuration
	logger.Init(logger.Config{
		Level:  logger.ParseLogLevel(cfg.LogLevel),
		Format: logger.ParseLogFormat(cfg.LogFormat),
	})

	logger.Info("Starting car service application")
	logger.Info("Environment: %s", cfg.Environment)
	logger.Debug("Configuration loaded successfully")

	// Connect to database
	db, err := database.NewDatabase("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	logger.Info("Database connection established successfully")

	// Initialize API server
	apiServer := server.NewAPIServer(db, cfg)
	logger.Info("API server initialized")

	// Start server
	logger.Info("Starting API server on port %s", cfg.Port)
	err = apiServer.Start(cfg.Port, cfg.AllowedOrigins)
	if err != nil {
		logger.Fatal("Failed to start API server: %v", err)
	}
}
