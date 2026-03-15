package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
)

type HouseholdHandler struct {
	svc domain.HouseholdService
}

func NewHouseholdHandler(svc domain.HouseholdService) *HouseholdHandler {
	return &HouseholdHandler{svc: svc}
}

func (h *HouseholdHandler) Create(c echo.Context) error {
	var input domain.CreateHouseholdInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if input.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if err := validateMaxLen("name", input.Name, 255); err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	household, err := h.svc.Create(c.Request().Context(), input, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create household")
	}

	return c.JSON(http.StatusCreated, toHouseholdResponse(household))
}

func (h *HouseholdHandler) Get(c echo.Context) error {
	id := c.Param("id")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	household, err := h.svc.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		return householdError(err)
	}

	return c.JSON(http.StatusOK, toHouseholdResponse(household))
}

func (h *HouseholdHandler) List(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	households, err := h.svc.ListByUser(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list households")
	}

	resp := make([]domain.HouseholdResponse, len(households))
	for i, hh := range households {
		resp[i] = toHouseholdResponse(&hh)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *HouseholdHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var input domain.UpdateHouseholdInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if input.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if err := validateMaxLen("name", input.Name, 255); err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	household, err := h.svc.Update(c.Request().Context(), id, input, userID)
	if err != nil {
		return householdError(err)
	}

	return c.JSON(http.StatusOK, toHouseholdResponse(household))
}

func (h *HouseholdHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if err := h.svc.Delete(c.Request().Context(), id, userID); err != nil {
		return householdError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *HouseholdHandler) Join(c echo.Context) error {
	var input domain.JoinHouseholdInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if input.InviteCode == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invite_code is required")
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	household, err := h.svc.Join(c.Request().Context(), input.InviteCode, userID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInviteCode) {
			return echo.NewHTTPError(http.StatusNotFound, "invalid invite code")
		}
		if errors.Is(err, domain.ErrAlreadyMember) {
			return echo.NewHTTPError(http.StatusConflict, "already a member")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to join household")
	}

	return c.JSON(http.StatusOK, toHouseholdResponse(household))
}

func (h *HouseholdHandler) ListMembers(c echo.Context) error {
	householdID := c.Param("id")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	members, err := h.svc.ListMembers(c.Request().Context(), householdID, userID)
	if err != nil {
		return householdError(err)
	}

	resp := make([]domain.MemberResponse, len(members))
	for i, m := range members {
		resp[i] = toMemberResponse(&m)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *HouseholdHandler) UpdateMemberSalary(c echo.Context) error {
	householdID := c.Param("id")
	memberID := c.Param("memberId")

	var input domain.UpdateSalaryInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if input.SalaryCents < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "salary_cents must be non-negative")
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	if err := h.svc.UpdateMemberSalary(c.Request().Context(), householdID, memberID, input.SalaryCents, userID); err != nil {
		return householdError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *HouseholdHandler) UpdateSplitMode(c echo.Context) error {
	householdID := c.Param("id")

	var input domain.UpdateSplitModeInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	if err := h.svc.UpdateSplitMode(c.Request().Context(), householdID, input.SplitMode, userID); err != nil {
		if errors.Is(err, domain.ErrInvalidSplitMode) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return householdError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *HouseholdHandler) UpdateMemberSplitPercentage(c echo.Context) error {
	householdID := c.Param("id")
	memberID := c.Param("memberId")

	var input domain.UpdateSplitPercentageInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	if err := h.svc.UpdateMemberSplitPercentage(c.Request().Context(), householdID, memberID, input.SplitPercentage, userID); err != nil {
		if errors.Is(err, domain.ErrInvalidSplitPercentage) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return householdError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *HouseholdHandler) RemoveMember(c echo.Context) error {
	householdID := c.Param("id")
	memberID := c.Param("memberId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if err := h.svc.RemoveMember(c.Request().Context(), householdID, memberID, userID); err != nil {
		return householdError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func RegisterHouseholdRoutes(e *echo.Echo, h *HouseholdHandler, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/households", authMiddleware)
	g.POST("", h.Create)
	g.GET("", h.List)
	g.POST("/join", h.Join)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	g.GET("/:id/members", h.ListMembers)
	g.PUT("/:id/members/:memberId/salary", h.UpdateMemberSalary)
	g.PUT("/:id/members/:memberId/percentage", h.UpdateMemberSplitPercentage)
	g.PUT("/:id/split-mode", h.UpdateSplitMode)
	g.DELETE("/:id/members/:memberId", h.RemoveMember)
}

func householdError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, domain.ErrHouseholdNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "household not found")
	case errors.Is(err, domain.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	case errors.Is(err, domain.ErrNotMember):
		return echo.NewHTTPError(http.StatusNotFound, "member not found")
	case errors.Is(err, domain.ErrAlreadyMember):
		return echo.NewHTTPError(http.StatusConflict, "already a member")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}

func toHouseholdResponse(h *domain.Household) domain.HouseholdResponse {
	return domain.HouseholdResponse{
		ID:         h.ID,
		Name:       h.Name,
		InviteCode: h.InviteCode,
		SplitMode:  h.SplitMode,
		CreatedAt:  h.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toMemberResponse(m *domain.HouseholdMember) domain.MemberResponse {
	return domain.MemberResponse{
		UserID:          m.UserID,
		UserName:        m.UserName,
		UserEmail:       m.UserEmail,
		SalaryCents:     m.SalaryCents,
		SplitPercentage: m.SplitPercentage,
		Role:            m.Role,
		JoinedAt:        m.JoinedAt.Format("2006-01-02T15:04:05Z"),
	}
}
