package main

import (
	"fmt"
	"log/slog"
	"net/http"
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
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using environment variables")
	}

	cfg := config.Load()

	if err := migrate.Run(cfg.DatabaseURL); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	db, err := config.NewDB(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	householdRepo := repository.NewHouseholdRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	fixedBillRepo := repository.NewFixedBillRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	summaryRepo := repository.NewSummaryRepository(db)
	snapshotRepo := repository.NewFixedBillSnapshotRepository(db)
	verificationRepo := repository.NewEmailVerificationRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	healthChecker := repository.NewHealthChecker(db)

	// Services
	tokenService := service.NewJWTTokenService(cfg.JWTSecret)
	emailService := service.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)
	authService := service.NewAuthService(userRepo, tokenService, verificationRepo, emailService, cfg.VerificationCodeTTL, passwordResetRepo, cfg.PasswordResetTTL, cfg.FrontendURL)
	householdService := service.NewHouseholdService(householdRepo)
	categoryService := service.NewCategoryService(categoryRepo, householdRepo)
	fixedBillService := service.NewFixedBillService(fixedBillRepo, householdRepo)
	snapshotService := service.NewFixedBillSnapshotService(snapshotRepo, householdRepo)
	expenseService := service.NewExpenseService(expenseRepo, householdRepo)
	summaryService := service.NewSummaryService(summaryRepo, householdRepo, expenseRepo, fixedBillRepo, snapshotRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	householdHandler := handler.NewHouseholdHandler(householdService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	fixedBillHandler := handler.NewFixedBillHandler(fixedBillService)
	snapshotHandler := handler.NewFixedBillSnapshotHandler(snapshotService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	summaryHandler := handler.NewSummaryHandler(summaryService)
	authMW := appMiddleware.JWTAuth(tokenService)

	// Router
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = customErrorHandler

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{cfg.FrontendURL},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	handler.RegisterHealthRoutes(e, healthChecker)
	handler.RegisterAuthRoutes(e, authHandler, authMW)
	handler.RegisterHouseholdRoutes(e, householdHandler, authMW)
	handler.RegisterCategoryRoutes(e, categoryHandler, authMW)
	handler.RegisterFixedBillRoutes(e, fixedBillHandler, authMW)
	handler.RegisterFixedBillSnapshotRoutes(e, snapshotHandler, authMW)
	handler.RegisterExpenseRoutes(e, expenseHandler, authMW)
	handler.RegisterSummaryRoutes(e, summaryHandler, authMW)

	// Serve frontend SPA if dist/ directory exists
	if _, err := os.Stat("dist"); err == nil {
		e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:   "dist",
			Index:  "index.html",
			HTML5:  true,
			Browse: false,
		}))
	}

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	slog.Info("Server starting", "port", port)
	if err := e.Start(":" + port); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func customErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	message := "internal server error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)
	}

	if code >= 500 {
		slog.Error("HTTP error", "status", code, "error", err)
	}

	_ = c.JSON(code, map[string]string{"error": message})
}
