package user_service

import (
	"compost-bin/service/middleware"
	"compost-bin/service/middleware/dao"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func RegisterUser(username, password string) error {
	hash := sha256.New()
	salt := rand.Text()
	password = base32.HexEncoding.EncodeToString(hash.Sum([]byte(password + salt)))

	user := dao.User{
		Username: username,
		Password: password,
		Salt:     salt,
	}
	res := middleware.GetDb().Create(&user)
	return parseDbError(res.Error)
}

func LoginWithEmail(email, password string) (dao.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := gorm.G[dao.User](middleware.GetDb()).Where("email=?", email).First(ctx)
	if err != nil {
		return user, parseDbError(err)
	}

	if user.Password != passwordHash(password, user.Salt) {
		return user, fmt.Errorf("密码错误！")
	}

	return user, nil
}

func LoginWithUsername(username, password string) (dao.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := gorm.G[dao.User](middleware.GetDb()).Where("username=?", username).First(ctx)
	if err != nil {
		return user, parseDbError(err)
	}

	if user.Password != passwordHash(password, user.Salt) {
		return user, fmt.Errorf("密码错误！")
	}

	return user, nil
}

// Profile returns all infomation about user whose id equals userId.
func Profile(userId int64) (dao.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := gorm.G[dao.User](middleware.GetDb()).Where("id=?", userId).First(ctx)
	return user, parseDbError(err)
}

func UpdatePwd2Fa(userId int64, pwd string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := gorm.G[dao.User](middleware.GetDb()).Where("id=?", userId).Update(ctx, "pwd_2fa", pwd)
	return parseDbError(err)
}

func Update2FaCertification(userId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := gorm.G[dao.User](middleware.GetDb()).Where("id=?", userId).Update(ctx, "tfa_certed", true)
	return parseDbError(err)
}

func parseDbError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("记录不存在")
	}

	var mySqlErr *mysql.MySQLError
	if errors.As(err, &mySqlErr) {
		if mySqlErr.Number == 1062 {
			if strings.Contains(mySqlErr.Message, "username") {
				return fmt.Errorf("用户名已存在")
			}
			if strings.Contains(mySqlErr.Message, "email") {
				return fmt.Errorf("邮箱已被注册")
			}
		}
	}

	return middleware.DatabaseFailure
}

func passwordHash(password, salt string) string {
	return base32.HexEncoding.EncodeToString(sha256.New().Sum([]byte(password + salt)))
}
