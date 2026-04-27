package handler

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

)

func (d *Dispatcher) broadcastFriendAdded(ctx *WSContext, aID, aName, bID, bName string) {
	onA := ctx.Chub.Clients[aID] != nil
	onB := ctx.Chub.Clients[bID] != nil
	if c := ctx.Chub.Clients[aID]; c != nil {
		_ = c.Conn.WriteJSON(FriendsListResponse{
			Type:    "friend_added",
			Success: true,
			Friend:  Friend{ID: bID, Username: bName, Online: onB},
		})
	}
	if c := ctx.Chub.Clients[bID]; c != nil {
		_ = c.Conn.WriteJSON(FriendsListResponse{
			Type:    "friend_added",
			Success: true,
			Friend:  Friend{ID: aID, Username: aName, Online: onA},
		})
	}
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

func validateAndGetClaims(tokenString string) (*MyCustomClaims, error) {
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


