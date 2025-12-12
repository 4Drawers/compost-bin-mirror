package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func RequestId() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rid := uuid.NewString()
			c.Set("Request-Id", rid)

			ctx := context.WithValue(c.Request().Context(), "Request-Id", rid)
			c.SetRequest(c.Request().WithContext(ctx))

			c.Response().Header().Set("X-Reqeust-Id", rid)

			return next(c)
		}
	}
}
