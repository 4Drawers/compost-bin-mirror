// This package provides standard return value for echo routers.
package echo_errors

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Success(ctx echo.Context, msg string, res any) error {
	return ctx.JSON(http.StatusOK, echo.Map{
		"code":   http.StatusOK,
		"msg":    msg,
		"result": res,
	})
}

func BadRequest(ctx echo.Context, msg string) error {
	return ctx.JSON(http.StatusBadRequest, echo.Map{
		"code":   http.StatusBadRequest,
		"msg":    msg,
		"result": nil,
	})
}

func LoginAgain(ctx echo.Context, msg string) error {
	return ctx.JSON(http.StatusUnauthorized, echo.Map{
		"code":   http.StatusUnauthorized,
		"msg":    msg,
		"result": nil,
	})
}

func Forbidden(ctx echo.Context, msg string) error {
	return ctx.JSON(http.StatusUnauthorized, echo.Map{
		"code":   http.StatusForbidden,
		"msg":    msg,
		"result": nil,
	})
}

func ServerBroken(ctx echo.Context, msg string) error {
	return ctx.JSON(http.StatusInternalServerError, echo.Map{
		"code":   http.StatusInternalServerError,
		"msg":    msg,
		"result": nil,
	})
}
