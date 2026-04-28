package webutil

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	
	"github.com/realgetOff/We_Plaid_Guilty/internal/db"
	"github.com/realgetOff/We_Plaid_Guilty/internal/metrics"

	"crypto/rand"
)

func HandleGuestAuth(c *gin.Context, dbs *db.DBSafe) {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	guestName := fmt.Sprintf("guest_%04d", n.Int64())
	var userID string
	var err error

	// SQL QUERIES
	userQuery :=	`
					INSERT INTO users (username, type)
					VALUES ($1, 'guest')
					RETURNING id;
					`
	
	profileQuery := `
					INSERT INTO profiles (id, display_name)
					VALUES ($1, $2)
					`

	err = db.DBQuery(dbs, userQuery, []any{guestName}, &userID)
	if err != nil {
		fmt.Println("Guest creation failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't create a guest user in the database."})
		return
	}
	fmt.Println("Guest name: " + guestName + " guest ID = " + userID)
	
	err = db.DBQuery(dbs, profileQuery, []any{userID, guestName})
	if err != nil {
		fmt.Println("Guest profile creation failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't create a guest profile in the database."})
		return
	}

	var SignedString string
	SignedString, err = GenerateJWT(userID, guestName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't sign / generate JWT for guest."})
		return
	}

	metrics.UserCountTotal.Inc()
	metrics.UserCountGuest.Inc()

	c.JSON(http.StatusOK, AuthResponse{
		Token: SignedString,
	})
}