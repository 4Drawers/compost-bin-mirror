package middleware

import (
	"compost-bin/logger"
	"time"

	"github.com/labstack/echo/v4"
)

func ReqeustTimeConsume() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			consume := time.Since(start)

			rid := c.Get("Request-Id")

			logger.Infof(
				"HTTP %s %s | status=%d | duration=%s | ip = %s | request-id=%v | user-agent=%s",
				c.Request().Method,
				c.Path(),
				c.Response().Status,
				consume,
				c.RealIP(),
				rid,
				c.Request().UserAgent(),
			)

			return err
		}
	}
}
