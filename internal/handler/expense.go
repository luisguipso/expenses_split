package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
)

type ExpenseHandler struct {
	svc domain.ExpenseService
}

func NewExpenseHandler(svc domain.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{svc: svc}
}

func (h *ExpenseHandler) Create(c echo.Context) error {
	householdID := c.Param("householdId")
	var input domain.CreateExpenseInput
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

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: creating expense",
		"household_id", householdID,
		"user_id", userID,
	)

	expense, err := h.svc.Create(c.Request().Context(), input, householdID, userID)
	if err != nil {
		slog.Error("handler: failed to create expense",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return expenseError(err)
	}

	slog.Info("handler: expense created",
		"expense_id", expense.ID,
		"household_id", householdID,
	)
	return c.JSON(http.StatusCreated, toExpenseResponse(expense))
}

func (h *ExpenseHandler) List(c echo.Context) error {
	householdID := c.Param("householdId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	filter := domain.ExpenseFilter{
		CategoryID: c.QueryParam("category_id"),
		UserID:     c.QueryParam("user_id"),
	}
	if m := c.QueryParam("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			filter.Month = v
		}
	}
	if y := c.QueryParam("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			filter.Year = v
		}
	}

	slog.Info("handler: listing expenses",
		"household_id", householdID,
		"user_id", userID,
	)

	expenses, err := h.svc.List(c.Request().Context(), householdID, userID, filter)
	if err != nil {
		slog.Error("handler: failed to list expenses",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return expenseError(err)
	}

	resp := make([]domain.ExpenseResponse, len(expenses))
	for i, e := range expenses {
		resp[i] = toExpenseResponse(&e)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *ExpenseHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var input domain.UpdateExpenseInput
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
	if input.ExpenseDate == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "expense_date is required")
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: updating expense",
		"expense_id", id,
		"user_id", userID,
	)

	expense, err := h.svc.Update(c.Request().Context(), id, input, userID)
	if err != nil {
		slog.Error("handler: failed to update expense",
			"error", err,
			"expense_id", id,
			"user_id", userID,
		)
		return expenseError(err)
	}

	slog.Info("handler: expense updated",
		"expense_id", expense.ID,
	)
	return c.JSON(http.StatusOK, toExpenseResponse(expense))
}

func (h *ExpenseHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: deleting expense",
		"expense_id", id,
		"user_id", userID,
	)

	if err := h.svc.Delete(c.Request().Context(), id, userID); err != nil {
		slog.Error("handler: failed to delete expense",
			"error", err,
			"expense_id", id,
			"user_id", userID,
		)
		return expenseError(err)
	}

	slog.Info("handler: expense deleted",
		"expense_id", id,
	)
	return c.NoContent(http.StatusNoContent)
}

func RegisterExpenseRoutes(e *echo.Echo, h *ExpenseHandler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/households/:householdId/expenses", authMiddleware)
	g.POST("", h.Create)
	g.GET("", h.List)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

func expenseError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, domain.ErrExpenseNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "expense not found")
	case errors.Is(err, domain.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}

func toExpenseResponse(e *domain.Expense) domain.ExpenseResponse {
	return domain.ExpenseResponse{
		ID:           e.ID,
		CategoryID:   e.CategoryID,
		CategoryName: e.CategoryName,
		Description:  e.Description,
		AmountCents:  e.AmountCents,
		ExpenseDate:  e.ExpenseDate,
		IsShared:     e.IsShared,
		PaidBy:       e.PaidBy,
		PaidByName:   e.PaidByName,
		AssignedTo:   e.AssignedTo,
	}
}
