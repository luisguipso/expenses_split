package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
)

type CategoryHandler struct {
	svc domain.CategoryService
}

func NewCategoryHandler(svc domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) Create(c echo.Context) error {
	householdID := c.Param("householdId")
	var input domain.CreateCategoryInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if input.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if err := validateMaxLen("name", input.Name, 100); err != nil {
		return err
	}
	if err := validateMaxLen("icon", input.Icon, 50); err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: creating category",
		"household_id", householdID,
		"user_id", userID,
	)

	cat, err := h.svc.Create(c.Request().Context(), input, householdID, userID)
	if err != nil {
		slog.Error("handler: failed to create category",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return categoryError(err)
	}

	slog.Info("handler: category created",
		"category_id", cat.ID,
		"household_id", householdID,
	)
	return c.JSON(http.StatusCreated, toCategoryResponse(cat))
}

func (h *CategoryHandler) List(c echo.Context) error {
	householdID := c.Param("householdId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: listing categories",
		"household_id", householdID,
		"user_id", userID,
	)

	categories, err := h.svc.List(c.Request().Context(), householdID, userID)
	if err != nil {
		slog.Error("handler: failed to list categories",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return categoryError(err)
	}

	resp := make([]domain.CategoryResponse, len(categories))
	for i, cat := range categories {
		resp[i] = toCategoryResponse(&cat)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *CategoryHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var input domain.UpdateCategoryInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if input.Name != "" {
		if err := validateMaxLen("name", input.Name, 100); err != nil {
			return err
		}
	}
	if err := validateMaxLen("icon", input.Icon, 50); err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: updating category",
		"category_id", id,
		"user_id", userID,
	)

	cat, err := h.svc.Update(c.Request().Context(), id, input, userID)
	if err != nil {
		slog.Error("handler: failed to update category",
			"error", err,
			"category_id", id,
			"user_id", userID,
		)
		return categoryError(err)
	}

	slog.Info("handler: category updated",
		"category_id", cat.ID,
	)
	return c.JSON(http.StatusOK, toCategoryResponse(cat))
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	slog.Info("handler: deleting category",
		"category_id", id,
		"user_id", userID,
	)

	if err := h.svc.Delete(c.Request().Context(), id, userID); err != nil {
		slog.Error("handler: failed to delete category",
			"error", err,
			"category_id", id,
			"user_id", userID,
		)
		return categoryError(err)
	}

	slog.Info("handler: category deleted",
		"category_id", id,
	)
	return c.NoContent(http.StatusNoContent)
}

func RegisterCategoryRoutes(e *echo.Echo, h *CategoryHandler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/households/:householdId/categories", authMiddleware)
	g.POST("", h.Create)
	g.GET("", h.List)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

func categoryError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, domain.ErrCategoryNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "category not found")
	case errors.Is(err, domain.ErrCategoryExists):
		return echo.NewHTTPError(http.StatusConflict, "category already exists")
	case errors.Is(err, domain.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}

func toCategoryResponse(c *domain.Category) domain.CategoryResponse {
	return domain.CategoryResponse{
		ID:   c.ID,
		Name: c.Name,
		Icon: c.Icon,
	}
}
