package util

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("")

var TokenExpireDuration time.Duration

type Claims struct {
	Address string `json:"address"`
	jwt.RegisteredClaims
}

func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

func SetTokenExpireDuration(duration time.Duration) {
	TokenExpireDuration = duration
}

func GenerateAuthToken(address string) (string, error) {
	claims := Claims{
		Address: address,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "owlto.finance",
			Subject:   address,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseAuthToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func ValidateToken(tokenString string) (bool, error) {
	claims, err := ParseAuthToken(tokenString)
	if err != nil {
		return false, err
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return false, errors.New("token expired")
	}

	return true, nil
}

func GetAddressFromToken(tokenString string) (string, error) {
	claims, err := ParseAuthToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Address, nil
}

func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseAuthToken(tokenString)
	if err != nil {
		return "", err
	}

	if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) < time.Hour*24 {
		return GenerateAuthToken(claims.Address)
	}

	return tokenString, nil
}
