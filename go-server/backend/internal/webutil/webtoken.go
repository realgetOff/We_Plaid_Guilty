package webutil

import (
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

var JwtSecret = []byte("replace_with_env_or_equivalent_later")

type MyCustomClaims struct {
	Username string `json:"username"`
	UserID   string `json:"id"`
	jwt.RegisteredClaims
}

type AuthResponse struct {
	Token string `json:"token"`
}

func GenerateJWT(userID string, guestName string) (string, error) {
	claims := MyCustomClaims{
		Username: guestName,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // change later, temporarily 24h
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(JwtSecret)
	if err != nil {
		fmt.Println("Couldn't sign / generate JWT for guest " + guestName + " where id = " + userID)
		return "", err
	}
	return signedToken, nil
}

func ValidateAndGetClaims(tokenString string) (*MyCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})

	if err != nil || token == nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("Token is invalid or claims are corrupted")
}


