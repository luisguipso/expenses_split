package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
)

type ImportHandler struct {
	svc domain.ImportService
}

func NewImportHandler(svc domain.ImportService) *ImportHandler {
	return &ImportHandler{svc: svc}
}

func (h *ImportHandler) Upload(c echo.Context) error {
	householdID := c.Param("householdId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	if file.Size > 10<<20 {
		return echo.NewHTTPError(http.StatusBadRequest, "file must be at most 10MB")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	ct := file.Header.Get("Content-Type")
	if ext != ".pdf" && ct != "application/pdf" {
		return echo.NewHTTPError(http.StatusBadRequest, "only PDF files are supported")
	}

	src, err := file.Open()
	if err != nil {
		slog.Error("failed to open uploaded file", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		slog.Error("failed to read uploaded file", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	resp, err := h.svc.ParseBill(c.Request().Context(), file.Filename, content, householdID, userID)
	if err != nil {
		return importError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *ImportHandler) Confirm(c echo.Context) error {
	householdID := c.Param("householdId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var input domain.ImportConfirmInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if len(input.Items) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "items cannot be empty")
	}

	expenses, err := h.svc.ConfirmImport(c.Request().Context(), input, householdID, userID)
	if err != nil {
		return importError(err)
	}

	resp := make([]domain.ExpenseResponse, len(expenses))
	for i, e := range expenses {
		resp[i] = toExpenseResponse(&e)
	}
	return c.JSON(http.StatusCreated, resp)
}

func RegisterImportRoutes(e *echo.Echo, h *ImportHandler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/households/:householdId/import", authMiddleware)
	g.POST("/upload", h.Upload)
	g.POST("/confirm", h.Confirm)
}

func importError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, domain.ErrUnsupportedBillFormat):
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported bill format")
	case errors.Is(err, domain.ErrBillParseError):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "failed to parse bill")
	case errors.Is(err, domain.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	default:
		slog.Error("import error", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}
