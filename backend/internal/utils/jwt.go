package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTProvider interface {
	GenerateAccessToken(userName string, expirationTime, issuedAt, notBefore time.Time) (string, error)
	GenerateRefreshToken(userName string, ext, iat, notBefore time.Time) (string, error)
	VerifyJWT(tokenString string) (*JWTClaims, error)
}

type JWTClaims struct {
	UserName string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type DefaultJWTProvider struct {
	secret string
	issuer string
}

func NewDefaultJWTProvider(jwtSecret, issuer string) *DefaultJWTProvider {
	return &DefaultJWTProvider{
		secret: jwtSecret,
		issuer: issuer,
	}
}

func (p *DefaultJWTProvider) GenerateAccessToken(userName, role string, ext, iat, notBefore time.Time) (string, error) {
	if p.secret == "" {
		return "", errors.New("jwt secret is required")
	}

	claims := &JWTClaims{
		UserName: userName,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(ext),
			IssuedAt:  jwt.NewNumericDate(iat),
			NotBefore: jwt.NewNumericDate(notBefore),
			Issuer:    p.issuer,
			Subject:   userName,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.secret))
}

func (p *DefaultJWTProvider) GenerateRefreshToken(userName, role string, ext, iat, notBefore time.Time) (string, error) {
	refreshToken, err := p.GenerateAccessToken(
		userName,
		role,
		ext,
		iat,
		notBefore,
	)

	if err != nil {
		return "", errors.New("failed to generate refresh token")
	}

	return refreshToken, nil
}

func (p *DefaultJWTProvider) VerifyJWT(tokenString string) (*JWTClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, errors.New("invalid signing method")
			}
			return []byte(p.secret), nil
		},
	)

	if err != nil {
		return nil, errors.New("failed to parse token")
	}

	claims, ok := parsedToken.Claims.(*JWTClaims)

	if !ok || !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
