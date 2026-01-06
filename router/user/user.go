package user

import (
	"compost-bin/logger"
	echo_errors "compost-bin/router/echo-errors"
	"compost-bin/router/middleware"
	"compost-bin/service/jwt"
	db "compost-bin/service/middleware"
	"compost-bin/service/middleware/dao"
	user_service "compost-bin/service/user"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func WithUserApiV1(g *echo.Group) {
	user := g.Group("/user")
	{
		user.POST("/register", register)
		user.POST("/login", login)
		user.GET("/profile/:user_id", profile, middleware.Auth(getEnv("JWT_PASSWORD", "")))
		user.PUT("/profile/email/:user_id", changeEmail, middleware.Auth(getEnv("JWT_PASSWORD", "")))
	}
}

func register(ctx echo.Context) error {
	username := ctx.FormValue("username")
	password := ctx.FormValue("password")
	if len([]rune(username)) > 20 {
		return echo_errors.BadRequest(ctx, "用户名不能超过20个字符！")
	}
	if err := user_service.RegisterUser(username, password); err != nil {
		return whoseProblem(ctx, err)
	}
	return echo_errors.Success(ctx, "注册成功！", nil)
}

func login(ctx echo.Context) error {
	userInfo := ctx.FormValue("user_info")
	password := ctx.FormValue("password")

	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	var user dao.User
	var err error
	if emailPattern.Match([]byte(userInfo)) {
		user, err = user_service.LoginWithEmail(userInfo, password)
	} else if len([]rune(userInfo)) <= 20 {
		user, err = user_service.LoginWithUsername(userInfo, password)
	} else {
		return echo_errors.BadRequest(ctx, "请输入正确的用户名/密码！")
	}
	if err != nil {
		return whoseProblem(ctx, err)
	}

	jwtBuilder := new(jwt.JwtBuilder)
	jwtBuilder.SetAccessSecret(getEnv("JWT_PASSWORD", "")).
		SetAccessExpire(time.Until(time.Now().Add(5 * time.Minute))).
		SetRefreshSecret(getEnv("JWT_PASSWORD", "")).
		SetRefreshExpire(time.Until(time.Now().Add(24 * time.Hour))).
		SetIssuer("composter-bin").
		SetSubject("User Login").
		SetAudience(user.Character).SetId(ctx.Get("Request-Id").(string))

	accessToken, err := jwt.GenerateTokens(ctx.Request().Context(), user.Id, true, func() *jwt.JwtBuilder {
		return jwtBuilder.SetClaim4AccessToken(user.Id)
	})
	if err != nil {
		logger.WithContex(ctx.Request().Context()).Errorf("Failed to generate access token: %v", err)
		return echo_errors.ServerBroken(ctx, "服务器错误！请联系管理员！")
	}

	refreshToken, err := jwt.GenerateTokens(ctx.Request().Context(), user.Id, false, func() *jwt.JwtBuilder {
		return jwtBuilder.SetClaim4RefreshToken(user.Id)
	})
	if err != nil {
		logger.WithContex(ctx.Request().Context()).Errorf("Failed to generate refresh token: %v", err)
		return echo_errors.ServerBroken(ctx, "服务器错误！请联系管理员！")
	}

	ctx.Response().Header().Set("X-Authorization", accessToken)
	ctx.Response().Header().Set("X-Refresh", refreshToken)
	return echo_errors.Success(ctx, "登录成功！", user.Id)
}

// profile allows users check their BASIC infomation about their accounts that's updated to
// the server by themselves.
// GET /user/profile/:user_id
func profile(ctx echo.Context) error {
	param := ctx.Param("user_id")

	userId, err := strconv.Atoi(param)
	if err != nil {
		return echo_errors.BadRequest(ctx, "请提供正确的用户ID！")
	}

	if ctx.Request().Header.Get("User-Id") != param {
		return echo_errors.BadRequest(ctx, "仅支持查询自己的账户信息！")
	}

	user, err := user_service.Profile(int64(userId))
	if err != nil {
		return echo_errors.BadRequest(ctx, err.Error())
	}

	result := make(map[string]any, 10)
	result["id"] = userId
	result["username"] = user.Username
	result["email"] = user.Email
	result["avatar"] = user.Avatar
	result["sign"] = user.Sign
	result["character"] = user.Character

	return echo_errors.Success(ctx, "查询成功", result)
}

// changeEmail lets users change the emails of themselves.
// - user hasn't registered an email address in profile of (him/her/...)self
//   - register this email address as user's unconfirmed email (redis cache)
//   - send an email to user's specified email address with a confirm code
//   - redirect to confirm page
//
// - user has registered an email address
//   - register this email address as user's unconfirmed email (redis cache)
//   - send an email to user's old email address with a confirm code
//   - redirect to confirm page
//
// PUT /user/profile/email/:user_id
func changeEmail(ctx echo.Context) error {
	// TODO: implement after email service.
	return nil
}

func getEnv(variableName, defaultValue string) string {
	if value := os.Getenv(variableName); value != "" {
		return value
	}
	return defaultValue
}

func whoseProblem(ctx echo.Context, err error) error {
	if err == db.DatabaseFailure {
		return echo_errors.ServerBroken(ctx, err.Error())
	}
	return echo_errors.BadRequest(ctx, err.Error())
}
