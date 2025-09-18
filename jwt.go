package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateJWT(playerID string) (string, error) {
	claims := jwt.MapClaims{
		"playerID": playerID,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func parseJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	playerID, ok := claims["playerID"].(string)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}
	return playerID, nil
}
