package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
)

type FixedBillSnapshotHandler struct {
	svc domain.FixedBillSnapshotService
}

func NewFixedBillSnapshotHandler(svc domain.FixedBillSnapshotService) *FixedBillSnapshotHandler {
	return &FixedBillSnapshotHandler{svc: svc}
}

func (h *FixedBillSnapshotHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var input domain.UpdateFixedBillSnapshotInput
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

	slog.Info("handler: updating fixed bill snapshot",
		"snapshot_id", id,
		"user_id", userID,
	)

	snap, err := h.svc.Update(c.Request().Context(), id, input, userID)
	if err != nil {
		slog.Error("handler: failed to update fixed bill snapshot",
			"error", err,
			"snapshot_id", id,
			"user_id", userID,
		)
		return snapshotError(err)
	}

	slog.Info("handler: fixed bill snapshot updated",
		"snapshot_id", snap.ID,
	)
	return c.JSON(http.StatusOK, domain.FixedBillSnapshotResponse{
		ID:          snap.ID,
		FixedBillID: snap.FixedBillID,
		CategoryID:  snap.CategoryID,
		Description: snap.Description,
		AmountCents: snap.AmountCents,
		DueDay:      snap.DueDay,
		IsShared:    snap.IsShared,
		PaidBy:      snap.PaidBy,
		AssignedTo:  snap.AssignedTo,
		IsFrozen:    true,
	})
}

func RegisterFixedBillSnapshotRoutes(e *echo.Echo, h *FixedBillSnapshotHandler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/households/:householdId/bills/snapshots", authMiddleware)
	g.PUT("/:id", h.Update)
}

func snapshotError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, domain.ErrFixedBillSnapshotNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "fixed bill snapshot not found")
	case errors.Is(err, domain.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}
