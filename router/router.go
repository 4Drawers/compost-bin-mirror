package router

import (
	"compost-bin/router/middleware"
	"compost-bin/router/user"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Latest returns router with latest web api.
func Latest() *echo.Echo {
	router := echo.New()
	router.Use(middleware.RequestId())
	router.Use(middleware.ReqeustTimeConsume())

	router.GET("/", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "Hello, world!")
	})

	withApiV1(router)

	return router
}

func withApiV1(router *echo.Echo) {
	v1 := router.Group("/v1")
	{
		user.WithUserApiV1(v1)
	}
}
