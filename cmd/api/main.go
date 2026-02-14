package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lguilherme/contas/internal/config"
	"github.com/lguilherme/contas/internal/handler"
	appMiddleware "github.com/lguilherme/contas/internal/middleware"
	"github.com/lguilherme/contas/internal/migrate"
	"github.com/lguilherme/contas/internal/repository"
	"github.com/lguilherme/contas/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()

	if err := migrate.Run(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	db, err := config.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	healthChecker := repository.NewHealthChecker(db)

	// Services
	tokenService := service.NewJWTTokenService(cfg.JWTSecret)
	authService := service.NewAuthService(userRepo, tokenService)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	authMW := appMiddleware.JWTAuth(tokenService)

	// Router
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{cfg.FrontendURL},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	handler.RegisterHealthRoutes(e, healthChecker)
	handler.RegisterAuthRoutes(e, authHandler, authMW)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
		os.Exit(1)
	}
}
