package handler

import (
	"errors"
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
	userID := c.Get("user_id").(string)

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

	summary, err := h.svc.Generate(c.Request().Context(), householdID, year, month, userID)
	if err != nil {
		return summaryError(err)
	}

	return c.JSON(http.StatusOK, summary)
}

func (h *SummaryHandler) GetDashboard(c echo.Context) error {
	householdID := c.Param("householdId")
	userID := c.Get("user_id").(string)

	dashboard, err := h.svc.GetDashboard(c.Request().Context(), householdID, userID)
	if err != nil {
		return summaryError(err)
	}

	return c.JSON(http.StatusOK, dashboard)
}

func RegisterSummaryRoutes(e *echo.Echo, h *SummaryHandler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/households/:householdId", authMiddleware)
	g.GET("/summary", h.GetSummary)
	g.GET("/dashboard", h.GetDashboard)
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
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}
