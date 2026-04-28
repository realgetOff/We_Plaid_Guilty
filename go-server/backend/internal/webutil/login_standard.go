package webutil

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	
	"github.com/realgetOff/We_Plaid_Guilty/internal/db"
	"github.com/realgetOff/We_Plaid_Guilty/internal/metrics"

	"golang.org/x/crypto/bcrypt"
)

func HandleRegister(c *gin.Context, dbs *db.DBSafe) {
	var login loginInfo
	var userID string

	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	lenUsrname := len(login.Username)

	if lenUsrname < 3 || lenUsrname > 16 {
		return
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(login.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't hash password: " + err.Error()})
		return
	}

	// SQL QUERIES
	userQuery :=	`
					INSERT INTO users (username, email, password_hash)
					VALUES ($1, $2, $3)
					RETURNING id;
					`

	profileQuery :=	`
					INSERT INTO profiles (id, display_name)
					VALUES ($1, $2)
					`
	
	err = db.DBQuery(dbs, userQuery, []any{login.Username, login.Email, bytes}, &userID)
	if err != nil {
		fmt.Println("User registration failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't register a user in the database."})
		return
	}

	fmt.Println("Registered username: " + login.Username + USR_ID + userID)

	err = db.DBQuery(dbs, profileQuery, []any{userID, login.Username})
	if err != nil {
		fmt.Println("User profile creation failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't create a user profile in the database."})
		return
	}

	var SignedString string
	SignedString, err = GenerateJWT(userID, login.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": JWT_ERROR})
		return
	}

	metrics.UserCountTotal.Inc()
	metrics.UserCountStandard.Inc()

	c.JSON(http.StatusOK, AuthResponse{
		Token: SignedString,
	})
}

func HandleLogin(c *gin.Context, dbs *db.DBSafe) {
	var login loginInfo
	var userID string
	var passHash string
	var err error

	if err = c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	userQuery :=	`
					SELECT id, password_hash
					FROM users
					WHERE username = $1 AND type = 'standard';
					`

	err = db.DBQuery(dbs, userQuery, []any{login.Username}, &userID, &passHash)

	if err != nil {
		fmt.Println("Coulnd't get password hash for user: "+login.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server get the password hash for the user."})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passHash), []byte(login.Password))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "The passwords don't match / the hash comparison failed."})
		return
	}

	fmt.Println("Password valid for: " + login.Username + USR_ID + userID)

	var SignedString string
	SignedString, err = GenerateJWT(userID, login.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": JWT_ERROR})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: SignedString,
	})
}
