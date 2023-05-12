// Package httperr provides some convenience functions to create echo HTTP errors.
package httperr

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Newf creates an HTTP error with a code and message.
func Newf(code int, format string, args ...interface{}) error {
	return echo.NewHTTPError(code, fmt.Sprintf(format, args...))
}

// NotFoundf creates an HTTP error with "not found" code and a message.
func NotFoundf(format string, args ...interface{}) error {
	return Newf(http.StatusNotFound, format, args...)
}

// BadRequestf creates an HTTP error with "bad request" code and a message.
func BadRequestf(format string, args ...interface{}) error {
	return Newf(http.StatusBadRequest, format, args...)
}

// Unauthorizedf creates an HTTP error with "unauthorized" code and a message.
func Unauthorizedf(format string, args ...interface{}) error {
	return Newf(http.StatusUnauthorized, format, args...)
}

// BindFail creates an HTTP error with "bind fail" message.
func BindFail(err error) error {
	return BadRequestf("cannot bind form: %v", err)
}
