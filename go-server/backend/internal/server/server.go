package server

import (
	"fmt"
	"log"
	"time"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/zsais/go-gin-prometheus"

	"crypto/tls"
	"crypto/x509"

	"github.com/realgetOff/We_Plaid_Guilty/internal/db"
	"github.com/realgetOff/We_Plaid_Guilty/internal/gamemanager"
	"github.com/realgetOff/We_Plaid_Guilty/internal/handler"
	"github.com/realgetOff/We_Plaid_Guilty/internal/config"

)

func NewServerStructure () *ServerVarsStruct {
	var dbs db.DBSafe

	dbPool, err := db.ConnectToDatabase()
	dbs.Pool = dbPool
	go db.ReloadConfig(&dbs)
	if err != nil {
		log.Fatalf("Couldn't connect to the PostgreSQL database: %v", err)
	}
	hub := &gamemanager.Hub{
		Rooms: make(map[string]gamemanager.GameRoom),
	}
	r := gin.Default()

	gin_prom := ginprometheus.NewPrometheus("app")
	gin_prom.Use(r)

	db.StartupUserMetrics(&dbs)

	chub := &handler.ClientHub{
		Clients:	make(map[string]*handler.Client),
		Db:			&dbs,
	}

	return &ServerVarsStruct{
		globalHub:		hub,
		router:			r,
		db:				&dbs,
		ClientHub:		chub,
	}
}

func HealthChecker() {
	healthRouter := gin.New()
	healthRouter.Use(gin.Recovery())

	healthRouter.GET("/health", health)
	if err := healthRouter.Run(":8081"); err != nil {
		log.Fatalf("Error: launch server health %v", err)
	}
}

func CloseDB(serverVars *ServerVarsStruct) {
	serverVars.db.GetPool().Close()
}

func LocalHost(serverVars *ServerVarsStruct) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("HOSTING TRANSCENDENCE LOCALLY...")
	if err := serverVars.router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)	
	}
}

func Host(serverVars *ServerVarsStruct) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	tlsContent := config.Addnewlinestotls()
	if tlsContent == nil {
		log.Fatalf("Failed to read TLS file")
	}
	serverCert, err := tls.X509KeyPair(tlsContent, tlsContent)
	if (err != nil){
		log.Fatalf("Failed to parse key pair: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(tlsContent)

	tlsConfig := &tls.Config {
		Certificates:	[]tls.Certificate{serverCert},
		ClientCAs:		caCertPool,
		ClientAuth:		tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr: ":" + port,
		Handler: serverVars.router,
		TLSConfig: tlsConfig,
		ReadTimeout: 5 * time.Second,
	}

	fmt.Println(" ~~ Attempting to boot with mTLS on port ", port, " ~~")

	if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to run server over mTLS: %v", err)
	}
}

