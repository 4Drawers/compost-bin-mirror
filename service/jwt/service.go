package jwt

import (
	"compost-bin/logger"
	"compost-bin/service/middleware"
	"context"
	"fmt"
	"time"
)

func GenerateTokens(ctx context.Context, userId int64, isAccess bool,
	builder func() *JwtBuilder) (token string, err error) {
	b := builder()
	if isAccess {
		token, err = b.BuildAccessToken()
	} else {
		token, err = b.BuildRefreshToken()
	}
	if err != nil {
		logger.WithContex(ctx).Errorf("Failed to sign jwt token: %v", err)
		return "", fmt.Errorf("触发安全机制，签名失败（T_T）")
	}

	key := fmt.Sprintf("token:access=%T:%d", isAccess, userId)
	err = middleware.GetCache().Set(ctx, key, token, b.refreshExpire).Err()
	if err != nil {
		logger.WithContex(ctx).Errorf("Failed to store jwt token to redis: %v", err)
		return "", fmt.Errorf("系统异常")
	}

	return
}

func ValidRefresh(ctx context.Context, userId int64, token string) bool {
	key := fmt.Sprintf("token:access=%T:%d", false, userId)
	stored, err := middleware.GetCache().Get(ctx, key).Result()
	if err != nil {
		logger.WithContex(ctx).Errorf("Failed to take jwt token %s from redis: %v", key, err)
		return false
	}
	return stored == token
}

func TryBlackListToken(ctx context.Context, userId int64) error {
	key := fmt.Sprintf("token:access=%T:%d", true, userId)
	if err := middleware.GetCache().Del(ctx, key).Err(); err != nil {
		logger.WithContex(ctx).Errorf("Failed to delete jwt token %s from redis: %v", key, err)
	}
	return fmt.Errorf("系统异常")
}

func BlackListToken(ctx context.Context, userId int64) {
	ticker := time.NewTicker(50 * time.Millisecond)
	for i := 0; i < 3; i++ {
		if TryBlackListToken(ctx, userId) == nil {
			break
		}
		<-ticker.C
	}
}

func TokenBlackListed(ctx context.Context, userId int64) bool {
	key := fmt.Sprintf("token:access=%T:%d", true, userId)
	if err := middleware.GetCache().Get(ctx, key).Err(); err != nil {
		logger.WithContex(ctx).Errorf("Failed to take jwt token %s from redis: %v", key, err)
		return true
	}
	return false
}
