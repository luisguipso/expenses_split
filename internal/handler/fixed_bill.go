package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
)

type FixedBillHandler struct {
	svc domain.FixedBillService
}

func NewFixedBillHandler(svc domain.FixedBillService) *FixedBillHandler {
	return &FixedBillHandler{svc: svc}
}

func (h *FixedBillHandler) Create(c echo.Context) error {
	householdID := c.Param("householdId")
	var input domain.CreateFixedBillInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if input.Description == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "description is required")
	}
	if input.AmountCents <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "amount_cents must be positive")
	}
	if input.DueDay < 1 || input.DueDay > 31 {
		return echo.NewHTTPError(http.StatusBadRequest, "due_day must be between 1 and 31")
	}

	userID := c.Get("user_id").(string)
	bill, err := h.svc.Create(c.Request().Context(), input, householdID, userID)
	if err != nil {
		return fixedBillError(err)
	}

	return c.JSON(http.StatusCreated, toFixedBillResponse(bill))
}

func (h *FixedBillHandler) List(c echo.Context) error {
	householdID := c.Param("householdId")
	userID := c.Get("user_id").(string)

	bills, err := h.svc.List(c.Request().Context(), householdID, userID)
	if err != nil {
		return fixedBillError(err)
	}

	resp := make([]domain.FixedBillResponse, len(bills))
	for i, b := range bills {
		resp[i] = toFixedBillResponse(&b)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *FixedBillHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var input domain.UpdateFixedBillInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if input.Description == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "description is required")
	}
	if input.AmountCents <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "amount_cents must be positive")
	}
	if input.DueDay < 1 || input.DueDay > 31 {
		return echo.NewHTTPError(http.StatusBadRequest, "due_day must be between 1 and 31")
	}

	userID := c.Get("user_id").(string)
	bill, err := h.svc.Update(c.Request().Context(), id, input, userID)
	if err != nil {
		return fixedBillError(err)
	}

	return c.JSON(http.StatusOK, toFixedBillResponse(bill))
}

func (h *FixedBillHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	userID := c.Get("user_id").(string)

	if err := h.svc.Delete(c.Request().Context(), id, userID); err != nil {
		return fixedBillError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func RegisterFixedBillRoutes(e *echo.Echo, h *FixedBillHandler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/households/:householdId/bills", authMiddleware)
	g.POST("", h.Create)
	g.GET("", h.List)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

func fixedBillError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, domain.ErrFixedBillNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "fixed bill not found")
	case errors.Is(err, domain.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}

func toFixedBillResponse(b *domain.FixedBill) domain.FixedBillResponse {
	return domain.FixedBillResponse{
		ID:           b.ID,
		CategoryID:   b.CategoryID,
		CategoryName: b.CategoryName,
		Description:  b.Description,
		AmountCents:  b.AmountCents,
		DueDay:       b.DueDay,
		IsShared:     b.IsShared,
		AssignedTo:   b.AssignedTo,
		IsActive:     b.IsActive,
	}
}
