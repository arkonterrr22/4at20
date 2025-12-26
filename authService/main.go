package main

import (
	"context"
	"log"
	"os"
	"time"

	"authService/db"
	"authService/server"
)

func main() {
	jwtSecret := "super-secret-pass"
	connStr := os.Getenv("POSTGRES_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	database, err := db.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	if err := database.RunMigrations(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations completed successfully")
	srv := server.New(database, jwtSecret)
	port := os.Getenv("SERVER_PORT")
	log.Printf("Auth Service starting on :%s", port)
	if err := srv.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
