package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"studytracker/internal/router"
)

func main() {
	if err := godotenv.Load(".env"); err == nil {
		log.Println("loaded .env")
	}
	if err := godotenv.Load("../.env"); err == nil {
		log.Println("loaded ../.env")
	}
	log.Printf("GOOGLE_CLIENT_ID=%s", os.Getenv("GOOGLE_CLIENT_ID"))
	log.Printf("GOOGLE_REDIRECT_URL=%s", os.Getenv("GOOGLE_REDIRECT_URL"))
	log.Printf("GOOGLE_CLIENT_SECRET length=%d", len(os.Getenv("GOOGLE_CLIENT_SECRET")))

	addr := ":" + getEnv("PORT", "8080")
	dsn := getEnv("DATABASE_URL", "file:data/studytracker.db?_pragma=foreign_keys(ON)")

	app, err := router.New(dsn)
	if err != nil {
		log.Fatalf("failed to bootstrap router: %v", err)
	}

	log.Printf("study-tracker API listening on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
