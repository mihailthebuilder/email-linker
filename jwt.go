package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	User User `json:"user"`
	jwt.RegisteredClaims
}

type User struct {
	Id string `json:"string"`
}

func newJwtClient() *JwtClient {
	return &JwtClient{Secret: getEnv("JWT_SECRET")}
}

type JwtClient struct {
	Secret string
}

func (j *JwtClient) GenerateToken(userId string) (string, error) {
	claims := Claims{}
	claims.User.Id = userId
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(10 * time.Minute))
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(j.Secret))
}

func (j *JwtClient) GetUserId(token string) (string, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.Secret), nil
	})

	if err != nil {
		return "", fmt.Errorf("error parsing token %s: %s", token, err)
	}

	if !parsed.Valid {
		return "", fmt.Errorf("invalid token %s", token)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return "", fmt.Errorf("invalid claims for token %s", token)
	}

	userId := claims.User.Id
	if lengthOfString(userId) == 0 {
		return "", fmt.Errorf("no User.Id claim for token %s", token)
	}

	return userId, err
}
