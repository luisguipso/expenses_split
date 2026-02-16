package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/config"
	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/handler"
	appMiddleware "github.com/lguilherme/contas/internal/middleware"
	"github.com/lguilherme/contas/internal/migrate"
	"github.com/lguilherme/contas/internal/repository"
	"github.com/lguilherme/contas/internal/service"
)

// testEnv holds the shared test environment.
type testEnv struct {
	db     *pgxpool.Pool
	server *httptest.Server
	echo   *echo.Echo
}

var env *testEnv

func TestMain(m *testing.M) {
	dbURL := os.Getenv("CONTAS_TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://contas:contas@localhost:5432/contas_test?sslmode=disable"
	}

	if err := migrate.Run(dbURL); err != nil {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}

	db, err := config.NewDB(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connect failed: %v\n", err)
		os.Exit(1)
	}

	e := setupEcho(db)
	ts := httptest.NewServer(e)

	env = &testEnv{db: db, server: ts, echo: e}

	code := m.Run()

	ts.Close()
	db.Close()
	os.Exit(code)
}

const jwtSecret = "integration-test-secret"

func setupEcho(db *pgxpool.Pool) *echo.Echo {
	userRepo := repository.NewUserRepository(db)
	householdRepo := repository.NewHouseholdRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	fixedBillRepo := repository.NewFixedBillRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	summaryRepo := repository.NewSummaryRepository(db)
	snapshotRepo := repository.NewFixedBillSnapshotRepository(db)
	healthChecker := repository.NewHealthChecker(db)

	tokenService := service.NewJWTTokenService(jwtSecret)
	authService := service.NewAuthService(userRepo, tokenService)
	householdService := service.NewHouseholdService(householdRepo)
	categoryService := service.NewCategoryService(categoryRepo, householdRepo)
	fixedBillService := service.NewFixedBillService(fixedBillRepo, householdRepo)
	snapshotService := service.NewFixedBillSnapshotService(snapshotRepo, householdRepo)
	expenseService := service.NewExpenseService(expenseRepo, householdRepo)
	summaryService := service.NewSummaryService(summaryRepo, householdRepo, expenseRepo, fixedBillRepo, snapshotRepo)

	authHandler := handler.NewAuthHandler(authService)
	householdHandler := handler.NewHouseholdHandler(householdService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	fixedBillHandler := handler.NewFixedBillHandler(fixedBillService)
	snapshotHandler := handler.NewFixedBillSnapshotHandler(snapshotService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	summaryHandler := handler.NewSummaryHandler(summaryService)
	authMW := appMiddleware.JWTAuth(tokenService)

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		code := http.StatusInternalServerError
		message := "internal server error"
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = fmt.Sprintf("%v", he.Message)
		}
		_ = c.JSON(code, map[string]string{"error": message})
	}

	handler.RegisterHealthRoutes(e, healthChecker)
	handler.RegisterAuthRoutes(e, authHandler, authMW)
	handler.RegisterHouseholdRoutes(e, householdHandler, authMW)
	handler.RegisterCategoryRoutes(e, categoryHandler, authMW)
	handler.RegisterFixedBillRoutes(e, fixedBillHandler, authMW)
	handler.RegisterFixedBillSnapshotRoutes(e, snapshotHandler, authMW)
	handler.RegisterExpenseRoutes(e, expenseHandler, authMW)
	handler.RegisterSummaryRoutes(e, summaryHandler, authMW)

	return e
}

// cleanDB truncates all application tables between tests.
func cleanDB(t *testing.T) {
	t.Helper()
	tables := []string{
		"fixed_bill_snapshots",
		"monthly_summary_items",
		"monthly_summaries",
		"expenses",
		"fixed_bills",
		"categories",
		"household_members",
		"households",
		"users",
	}
	for _, table := range tables {
		if _, err := env.db.Exec(context.Background(), "DELETE FROM "+table); err != nil {
			t.Fatalf("clean %s: %v", table, err)
		}
	}
}

// authUser holds credentials and tokens for a registered test user.
type authUser struct {
	ID          string
	Name        string
	Email       string
	AccessToken string
}

// registerUser creates a new user via the API and returns auth info.
func registerUser(t *testing.T, name, email, password string) authUser {
	t.Helper()
	body := domain.RegisterInput{Name: name, Email: email, Password: password}
	resp := doJSON(t, http.MethodPost, "/auth/register", body, "", http.StatusCreated)

	var result domain.AuthResponse
	decodeJSON(t, resp, &result)

	return authUser{
		ID:          fmt.Sprintf("%v", result.User.ID),
		Name:        result.User.Name,
		Email:       result.User.Email,
		AccessToken: result.Tokens.AccessToken,
	}
}

// createHousehold creates a household via the API and returns its ID.
func createHousehold(t *testing.T, token, name string) string {
	t.Helper()
	body := domain.CreateHouseholdInput{Name: name}
	resp := doJSON(t, http.MethodPost, "/households", body, token, http.StatusCreated)

	var result map[string]interface{}
	decodeJSON(t, resp, &result)
	return result["id"].(string)
}

// joinHousehold makes a user join a household via invite code.
func joinHousehold(t *testing.T, joinerToken, householdID, ownerToken string) {
	t.Helper()
	resp := doGet(t, "/households/"+householdID, ownerToken, http.StatusOK)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	inviteCode := hh["invite_code"].(string)

	doJSON(t, http.MethodPost, "/households/join",
		domain.JoinHouseholdInput{InviteCode: inviteCode},
		joinerToken, http.StatusOK)
}

// doJSON sends a JSON request to the test server and asserts the expected status code.
func doJSON(t *testing.T, method, path string, body interface{}, token string, expectedStatus int) *http.Response {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, env.server.URL+path, reqBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Fatalf("%s %s: expected status %d, got %d", method, path, expectedStatus, resp.StatusCode)
	}
	return resp
}

// doGet is a convenience wrapper for GET requests.
func doGet(t *testing.T, path, token string, expectedStatus int) *http.Response {
	t.Helper()
	return doJSON(t, http.MethodGet, path, nil, token, expectedStatus)
}

// decodeJSON decodes a response body into the target struct.
func decodeJSON(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func currentYear() int  { return time.Now().Year() }
func currentMonth() int { return int(time.Now().Month()) }
