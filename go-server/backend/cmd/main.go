package main

import (
	"fmt"
	"os"

	"github.com/realgetOff/We_Plaid_Guilty/internal/server"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")
			
	serverVars := server.NewServerStructure()

	defer server.CloseDB(serverVars)
	
	gin.SetMode(gin.ReleaseMode)

	server.Routing(serverVars)
	go server.HealthChecker()

	if (os.Getenv("LOCAL") != "") {
		server.LocalHost(serverVars)
	} else {
		server.Host(serverVars)
	}		
}
