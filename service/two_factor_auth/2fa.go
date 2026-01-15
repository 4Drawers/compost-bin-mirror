package two_factor_auth_service

import (
	"fmt"
	"net/url"

	"github.com/pquerna/otp/totp"
)

func Generate(username string) (password string, encodedUrl string, err error) {
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      "堆肥桶",
		AccountName: username,
	})

	if password = Encrypt(key.Secret()); password == "" {
		return "", "", fmt.Errorf("服务器错误：无法生成 2FA 密钥，请联系管理员！")
	}

	encodedUrl = url.QueryEscape(key.URL())

	return password, encodedUrl, nil
}

func Certificate(password, code string) bool {
	return totp.Validate(code, Decrypt(password))
}
