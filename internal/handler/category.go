package handler

import (
	"errors"
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
	cat, err := h.svc.Create(c.Request().Context(), input, householdID, userID)
	if err != nil {
		return categoryError(err)
	}

	return c.JSON(http.StatusCreated, toCategoryResponse(cat))
}

func (h *CategoryHandler) List(c echo.Context) error {
	householdID := c.Param("householdId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	categories, err := h.svc.List(c.Request().Context(), householdID, userID)
	if err != nil {
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
	cat, err := h.svc.Update(c.Request().Context(), id, input, userID)
	if err != nil {
		return categoryError(err)
	}

	return c.JSON(http.StatusOK, toCategoryResponse(cat))
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if err := h.svc.Delete(c.Request().Context(), id, userID); err != nil {
		return categoryError(err)
	}

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
