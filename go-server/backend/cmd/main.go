package main

import (
	"fmt"
	"os"

	// "github.com/realgetOff/We_Plaid_Guilty/internal/handler"
	// "github.com/realgetOff/We_Plaid_Guilty/internal/config"
	// "github.com/realgetOff/We_Plaid_Guilty/internal/metrics"
	"github.com/realgetOff/We_Plaid_Guilty/internal/server"
	// "github.com/realgetOff/We_Plaid_Guilty/internal/webutil"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")
			
	//runtime.NumGoroutine()

	serverVars := server.NewServerStructure()

	defer server.CloseDB(serverVars)
	
	gin.SetMode(gin.ReleaseMode)
	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies 


	server.Routing(serverVars)
	go server.HealthChecker()

	if (os.Getenv("LOCAL") != "") {
		server.LocalHost(serverVars)
	} else {
		server.Host(serverVars)
	}		
}
