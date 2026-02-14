package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func getUserID(c echo.Context) (string, error) {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	return userID, nil
}

func validateMaxLen(field, value string, max int) error {
	if len(value) > max {
		return echo.NewHTTPError(http.StatusBadRequest, field+" must be at most "+strconv.Itoa(max)+" characters")
	}
	return nil
}

func isValidEmail(email string) bool {
	at := strings.Index(email, "@")
	if at < 1 {
		return false
	}
	domain := email[at+1:]
	dot := strings.LastIndex(domain, ".")
	return dot > 0 && dot < len(domain)-1
}
