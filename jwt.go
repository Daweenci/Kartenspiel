package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateJWT(playerID string) (string, error) {
	claims := jwt.MapClaims{
		"playerID": playerID,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}

func parseJWT(tokenString string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	playerID, ok := claims["playerID"].(string)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	return playerID, nil
}
