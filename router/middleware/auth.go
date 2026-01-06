package middleware

import (
	"compost-bin/logger"
	echo_errors "compost-bin/router/echo-errors"
	"compost-bin/service/jwt"
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func Auth(accessSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				return echo_errors.Forbidden(c, "请先登录")
			}

			token := strings.TrimPrefix(auth, "Bearer ")
			claims, err := jwt.ParseToken(token, accessSecret, true)
			if err != nil {
				return echo_errors.Forbidden(c, "数据无价，请您自重！")
			}

			refresher := func() *jwt.JwtBuilder {
				jwtBuilder := new(jwt.JwtBuilder)
				jwtBuilder.SetAccessSecret(accessSecret).
					SetAccessExpire(time.Until(time.Now().Add(5 * time.Minute))).
					SetIssuer("composter-bin").
					SetSubject("User Refresh").
					SetAudience(claims.Audience[0]).SetId(c.Get("Request-Id").(string))
				return jwtBuilder.SetClaim4AccessToken(claims.UserId)
			}

			if jwt.TokenBlackListed(c.Request().Context(), claims.UserId) {
				refTk := c.Request().Header.Get("Refresh")
				if !strings.HasPrefix(refTk, "Bearer ") ||
					!jwt.ValidRefresh(c.Request().Context(), claims.UserId, strings.TrimPrefix(refTk, "Bearer ")) {
					return echo_errors.LoginAgain(c, "距上次登录时间过长，请重新登录！")
				}

				token, err := jwt.GenerateTokens(c.Request().Context(), claims.UserId, true, refresher)
				if err != nil {
					logger.WithContex(c.Request().Context()).Errorf("Failed to refresh access token: %v", err)
					return echo_errors.ServerBroken(c, "安全系统崩溃，请联系管理员！")
				}
				c.Response().Header().Set("X-Authorization", token)
			} else if time.Now().Add(2*time.Minute).Compare(claims.ExpiresAt.Time) > 0 {
				token, err := jwt.GenerateTokens(c.Request().Context(), claims.UserId, true, refresher)
				if err != nil {
					logger.WithContex(c.Request().Context()).Errorf("Failed to refresh access token: %v", err)
				} else {
					jwt.BlackListToken(c.Request().Context(), claims.UserId)
					c.Response().Header().Set("X-Authorization", token)
				}
			}

			c.Request().Header.Add("User-Id", fmt.Sprintf("%d", claims.UserId))
			return next(c)
		}
	}
}
