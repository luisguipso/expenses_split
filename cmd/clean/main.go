package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/lguilherme/contas/internal/config"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	db, err := config.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	log.Println("🗑️  Cleaning database...")

	// Truncate all tables in reverse order of dependencies
	// CASCADE will handle dependent records automatically
	tables := []string{
		"monthly_summaries",
		"expenses",
		"fixed_bills",
		"categories",
		"household_members",
		"households",
		"users",
	}

	for _, table := range tables {
		log.Printf("Truncating %s...", table)
		_, err := db.Exec(ctx, "TRUNCATE TABLE "+table+" CASCADE")
		if err != nil {
			log.Fatalf("Failed to truncate %s: %v", table, err)
		}
	}

	log.Println("✅ Database cleaned successfully!")
}
