package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
)

type SummaryHandler struct {
	svc domain.SummaryService
}

func NewSummaryHandler(svc domain.SummaryService) *SummaryHandler {
	return &SummaryHandler{svc: svc}
}

func (h *SummaryHandler) GetSummary(c echo.Context) error {
	householdID := c.Param("householdId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	yearStr := c.QueryParam("year")
	monthStr := c.QueryParam("month")
	if yearStr == "" || monthStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "year and month query params are required")
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid year")
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid month")
	}

	slog.Info("handler: getting summary",
		"household_id", householdID,
		"user_id", userID,
		"year", year,
		"month", month,
	)

	summary, err := h.svc.Generate(c.Request().Context(), householdID, year, month, userID)
	if err != nil {
		slog.Error("handler: failed to get summary",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
			"year", year,
			"month", month,
		)
		return summaryError(err)
	}

	return c.JSON(http.StatusOK, summary)
}

func (h *SummaryHandler) GetDashboard(c echo.Context) error {
	householdID := c.Param("householdId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: getting dashboard",
		"household_id", householdID,
		"user_id", userID,
	)

	dashboard, err := h.svc.GetDashboard(c.Request().Context(), householdID, userID)
	if err != nil {
		slog.Error("handler: failed to get dashboard",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return summaryError(err)
	}

	return c.JSON(http.StatusOK, dashboard)
}

func (h *SummaryHandler) GetSummaryDetail(c echo.Context) error {
	householdID := c.Param("householdId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	targetUserID := c.QueryParam("user_id")
	if targetUserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id query param is required")
	}

	yearStr := c.QueryParam("year")
	monthStr := c.QueryParam("month")
	if yearStr == "" || monthStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "year and month query params are required")
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid year")
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid month")
	}

	slog.Info("handler: getting summary detail",
		"household_id", householdID,
		"user_id", userID,
		"target_user_id", targetUserID,
		"year", year,
		"month", month,
	)

	detail, err := h.svc.GetUserDetail(c.Request().Context(), householdID, year, month, targetUserID, userID)
	if err != nil {
		slog.Error("handler: failed to get summary detail",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
			"target_user_id", targetUserID,
			"year", year,
			"month", month,
		)
		return summaryError(err)
	}

	return c.JSON(http.StatusOK, detail)
}

func RegisterSummaryRoutes(e *echo.Echo, h *SummaryHandler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/households/:householdId/summary", authMiddleware)
	g.GET("", h.GetSummary)
	g.GET("/detail", h.GetSummaryDetail)

	d := e.Group("/households/:householdId/dashboard", authMiddleware)
	d.GET("", h.GetDashboard)
}

func summaryError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	case errors.Is(err, domain.ErrNoMembersWithSalary):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "no members have salary configured")
	case errors.Is(err, domain.ErrHouseholdNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "household not found")
	default:
		slog.Error("summary handler error", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}
