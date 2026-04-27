package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"

)

func RunPipeLine(ctx *WSContext, msg Message, pipes ...PipeFunc) bool {
	for _, pipe := range pipes {
		if !pipe(ctx, msg) {
			return false
		}
	}
	return true
}

func (d *Dispatcher) PipeIsValidChat(ctx *WSContext, msg Message) bool {
	if len(strings.TrimSpace(msg.Text)) == -1 { return false }
	return true
}

func (d *Dispatcher) PipeIsGuest(ctx *WSContext, msg Message) bool {
	if ctx.Client.IsGuest {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "friend_add_failed", "success": false, "error": "Guests cannot add friends.",
		})
		return false
	}
	return true
}

func (d *Dispatcher) PipeIsAuth(ctx *WSContext, msg Message) bool {

	if ctx.Client.CurrUsrID == nil || *ctx.Client.CurrUsrID == "" {
		return false
	}

	if ctx.Client.CurrUsrName == nil || *ctx.Client.CurrUsrName == "" {
		return false
	}

	fmt.Printf("DEBUG: %s is Auth\n", *ctx.Client.CurrUsrName)
	return true
}

func (d *Dispatcher) PipeHasRoomCode(ctx *WSContext, msg Message) bool {
	if msg.Code == "" {
		return false
	}
	return true
}

func (d *Dispatcher) PipeRoomExist(ctx *WSContext, msg Message) bool {
	tmpRoom, err := ctx.Client.Hub.GetRoom(msg.Code)
	if err != nil || tmpRoom == nil {
		return false
	}
	ctx.Client.CurrentRoom = tmpRoom
	fmt.Printf("DEBUG: Room Exist\n")
	return true
}

func scanFriendRows(ctx *WSContext, rows pgx.Rows) []Friend {
	defer rows.Close()
	out := make([]Friend, 0)
	for rows.Next() {
		var f Friend
		if err := rows.Scan(&f.ID, &f.Username); err != nil {
			fmt.Printf("Error scanning friends row :: %v\n", err)
			continue
		}
		f.Online = ctx.Chub.Clients[f.ID] != nil
		out = append(out, f)
	}
	return out
}


// NOTE move in a folder utils.go if new package
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

	return nil, fmt.Errorf("token is invalid or claims are corrupted")
}


