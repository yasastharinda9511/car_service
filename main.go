package main

import (
	"car_service/config"
	"car_service/database"
	"car_service/server"
	"log"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	database, err := database.NewDatabase("postgres", cfg.DatabaseURL)

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	apiServer := server.NewAPIServer(database)

	err = apiServer.Start(cfg.Port, cfg.AllowedOrigins)
	if err != nil {
		log.Fatal("Failed to start API server:", err)
	}
}
