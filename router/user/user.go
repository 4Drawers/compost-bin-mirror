package user

import (
	"compost-bin/logger"
	"compost-bin/router/middleware"
	"compost-bin/service/jwt"
	user_service "compost-bin/service/user"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func WithUserApiV1(g *echo.Group) {
	user := g.Group("/user")
	{
		user.POST("/register", register)
		user.POST("/login", login)
		user.GET("/info/:user_id", info, middleware.Auth(getEnv("JWT_PASSWORD", "")))
	}
}

func register(ctx echo.Context) error {
	username := ctx.FormValue("username")
	password := ctx.FormValue("password")
	err := user_service.RegisterUser(username, password)
	if err == nil {
		return ctx.JSON(http.StatusOK, "注册成功！")
	}
	return ctx.JSON(http.StatusBadRequest, fmt.Sprintf("注册失败：%v", err))
}

func login(ctx echo.Context) error {
	userInfo := ctx.FormValue("user_info")
	password := ctx.FormValue("password")
	user, err := user_service.Login(userInfo, password)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"code":   http.StatusBadRequest,
			"msg":    fmt.Sprintf("登录失败：%v", err),
			"result": nil,
		})
	}

	jwtBuilder := new(jwt.JwtBuilder)
	jwtBuilder.SetAccessSecret(getEnv("JWT_PASSWORD", "jwt_password")).
		SetAccessExpire(time.Until(time.Now().Add(5 * time.Minute))).
		SetRefreshSecret(getEnv("JWT_PASSWORD", "jwt_password")).
		SetRefreshExpire(time.Until(time.Now().Add(24 * time.Hour))).
		SetIssuer("composter-bin").
		SetSubject("User Login").
		SetAudience(user.Character).SetId(ctx.Get("Request-Id").(string))

	accessToken, err := jwt.GenerateTokens(ctx.Request().Context(), user.Id, true, func() *jwt.JwtBuilder {
		return jwtBuilder.SetClaim4AccessToken(user.Id)
	})
	if err != nil {
		logger.WithContex(ctx.Request().Context()).Errorf("Failed to generate access token: %v", err)
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"code":   http.StatusInternalServerError,
			"msg":    "服务器错误！请联系管理员！",
			"result": nil,
		})
	}
	refreshToken, err := jwt.GenerateTokens(ctx.Request().Context(), user.Id, false, func() *jwt.JwtBuilder {
		return jwtBuilder.SetClaim4RefreshToken(user.Id)
	})
	if err != nil {
		logger.WithContex(ctx.Request().Context()).Errorf("Failed to generate refresh token: %v", err)
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"code":   http.StatusInternalServerError,
			"msg":    "服务器错误！请联系管理员！",
			"result": nil,
		})
	}

	ctx.Response().Header().Set("X-Authorization", accessToken)
	ctx.Response().Header().Set("X-Refresh", refreshToken)
	return ctx.JSON(
		http.StatusOK,
		echo.Map{
			"code":   http.StatusOK,
			"msg":    "登录成功！",
			"result": user.Id,
		},
	)
}

// GET /user/info/:user_id
// Response:
// Content-Type: application/json
// All publication infomation of user with specified id.
func info(ctx echo.Context) error {
	arg := ctx.Param("user_id")

	id, err := strconv.Atoi(arg)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"code": http.StatusBadRequest,
			"msg":  "错误ID序号",
		})
	}

	// Get basic infomation
	user, err := user_service.PubInfo(int64(id))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"code": http.StatusBadRequest,
			"msg":  err.Error(),
		})
	}

	// TODO: get other infomation

	// edit result
	result := make(map[string]any, 10)
	result["username"] = user.Username
	result["avatar"] = user.Avatar
	result["sign"] = user.Sign
	result["character"] = user.Character
	if ctx.Request().Header.Get("User-Id") == arg {
		result["email"] = user.Email
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"code":   http.StatusOK,
		"msg":    "",
		"result": result,
	})
}

func getEnv(variableName, defaultValue string) string {
	if value := os.Getenv(variableName); value != "" {
		return value
	}
	return defaultValue
}
