package handler

import (
	"errors"
	"log/slog"
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
	if err := validateMaxLen("description", input.Description, 255); err != nil {
		return err
	}
	if input.AmountCents <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "amount_cents must be positive")
	}
	if input.DueDay < 1 || input.DueDay > 31 {
		return echo.NewHTTPError(http.StatusBadRequest, "due_day must be between 1 and 31")
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: creating fixed bill",
		"household_id", householdID,
		"user_id", userID,
	)

	bill, err := h.svc.Create(c.Request().Context(), input, householdID, userID)
	if err != nil {
		slog.Error("handler: failed to create fixed bill",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return fixedBillError(err)
	}

	slog.Info("handler: fixed bill created",
		"fixed_bill_id", bill.ID,
		"household_id", householdID,
	)
	return c.JSON(http.StatusCreated, toFixedBillResponse(bill))
}

func (h *FixedBillHandler) List(c echo.Context) error {
	householdID := c.Param("householdId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: listing fixed bills",
		"household_id", householdID,
		"user_id", userID,
	)

	bills, err := h.svc.List(c.Request().Context(), householdID, userID)
	if err != nil {
		slog.Error("handler: failed to list fixed bills",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
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
	if err := validateMaxLen("description", input.Description, 255); err != nil {
		return err
	}
	if input.AmountCents <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "amount_cents must be positive")
	}
	if input.DueDay < 1 || input.DueDay > 31 {
		return echo.NewHTTPError(http.StatusBadRequest, "due_day must be between 1 and 31")
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: updating fixed bill",
		"fixed_bill_id", id,
		"user_id", userID,
	)

	bill, err := h.svc.Update(c.Request().Context(), id, input, userID)
	if err != nil {
		slog.Error("handler: failed to update fixed bill",
			"error", err,
			"fixed_bill_id", id,
			"user_id", userID,
		)
		return fixedBillError(err)
	}

	slog.Info("handler: fixed bill updated",
		"fixed_bill_id", bill.ID,
	)
	return c.JSON(http.StatusOK, toFixedBillResponse(bill))
}

func (h *FixedBillHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: deleting fixed bill",
		"fixed_bill_id", id,
		"user_id", userID,
	)

	if err := h.svc.Delete(c.Request().Context(), id, userID); err != nil {
		slog.Error("handler: failed to delete fixed bill",
			"error", err,
			"fixed_bill_id", id,
			"user_id", userID,
		)
		return fixedBillError(err)
	}

	slog.Info("handler: fixed bill deleted",
		"fixed_bill_id", id,
	)
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
		PaidBy:       b.PaidBy,
		PaidByName:   b.PaidByName,
		AssignedTo:   b.AssignedTo,
		IsActive:     b.IsActive,
	}
}
