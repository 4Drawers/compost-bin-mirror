package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtBuilder struct {
	accessSecret  string
	refreshSecret string
	accessExpire  time.Duration
	refreshExpire time.Duration
	issuer        string
	subject       string
	audience      string
	id            string
	claims        Claims
}

type Claims struct {
	UserId int64 `json:"u_id"`
	jwt.RegisteredClaims
}

func (jb *JwtBuilder) BuildAccessToken() (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jb.claims).SignedString([]byte(jb.accessSecret))
}

func (jb *JwtBuilder) BuildRefreshToken() (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jb.claims).SignedString([]byte(jb.refreshSecret))
}

func ParseToken(tokenEncoded, secret string, isAccess bool) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenEncoded,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	} else {
		return claims, nil
	}
}

func (jb *JwtBuilder) SetClaim4AccessToken(userId int64) *JwtBuilder {
	jb.claims = Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jb.issuer,
			Subject:   jb.subject,
			Audience:  jwt.ClaimStrings{jb.audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jb.accessExpire)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jb.id,
		},
	}
	return jb
}

func (jb *JwtBuilder) SetClaim4RefreshToken(userId int64) *JwtBuilder {
	jb.claims = Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jb.issuer,
			Subject:   jb.subject,
			Audience:  jwt.ClaimStrings{jb.audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jb.refreshExpire)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jb.id,
		},
	}
	return jb
}

func (jb *JwtBuilder) SetAccessSecret(secret string) *JwtBuilder {
	jb.accessSecret = secret
	return jb
}

func (jb *JwtBuilder) SetRefreshSecret(secret string) *JwtBuilder {
	jb.refreshSecret = secret
	return jb
}

func (jb *JwtBuilder) SetAccessExpire(expire time.Duration) *JwtBuilder {
	jb.accessExpire = expire
	return jb
}

func (jb *JwtBuilder) SetRefreshExpire(expire time.Duration) *JwtBuilder {
	jb.refreshExpire = expire
	return jb
}

func (jb *JwtBuilder) SetIssuer(issuer string) *JwtBuilder {
	jb.issuer = issuer
	return jb
}

func (jb *JwtBuilder) SetSubject(subject string) *JwtBuilder {
	jb.subject = subject
	return jb
}

func (jb *JwtBuilder) SetAudience(audience string) *JwtBuilder {
	jb.audience = audience
	return jb
}

func (jb *JwtBuilder) SetId(id string) *JwtBuilder {
	jb.id = id
	return jb
}
