package main

import (
	"fmt"
	"log"
	"net/http" 
	"os"
	// "runtime"
	
	"main.go/gamemanager"
	"main.go/metrics"

	"github.com/zsais/go-gin-prometheus"
	"github.com/gin-gonic/gin"

	"crypto/tls"
	"crypto/x509"

	"golang.org/x/oauth2"
)


type serverVarsStruct struct {
	globalHub *gamemanager.Hub
	ClientHub *ClientHub
	router *gin.Engine
	db DBSafe
}

func NewServerStructure () *serverVarsStruct {
	var dbs DBSafe

	db, err := connectToDatabase()
	
	dbs.Pool = db
	go reloadConfig(&dbs)
	if err != nil {
		log.Fatalf("Couldn't connect to the PostgreSQL database: %v", err)
	}
	defer db.Close()

	dbPool, err := connectToDatabase()
	dbs.Pool = dbPool
	go reloadConfig(&dbs)
	if err != nil {
		log.Fatalf("Couldn't connect to the PostgreSQL database: %v", err)
	}
	hub := &gamemanager.Hub{
		Rooms: make(map[string]gamemanager.GameRoom),
	}
	r := gin.Default()

	gin_prom := ginprometheus.NewPrometheus("app")
	gin_prom.Use(r)
	metrics.RegisterMetrics()

	chub := &ClientHub{
		Clients:	make(map[string]*Client),
		db:			dbPool,
	}

	return &serverVarsStruct{
		globalHub:		hub,
		router:			r,
		db:				dbs,
		ClientHub:		chub,
	}
}

// https://api.intra.42.fr/apidoc/guides/web_application_flow#exchange-your-code-for-an-access-token
// https://pkg.go.dev/golang.org/x/oauth2#Endpoint

var (
	fortyTwoOauthConfig *oauth2.Config
	// this should be turned into a randomly generated string
	oauthStateString = "pseudo-random-state"
)

func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")


	// to be moved, temporarily here to allow docker-compose testing entirely locally
	if (os.Getenv("LOCAL") != "") {
		fortyTwoOauthConfig = &oauth2.Config {
			RedirectURL: "http://localhost:8080/api/auth/42/callback",
			ClientID: "u-s4t2ud-66ea3626f19c1907a4eb8c1c02ab89204b4cd2bf9d50d55f75f34464680e95a8",
			ClientSecret: "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",
			Scopes: []string{"public"},
			Endpoint:	oauth2.Endpoint {
				AuthURL: "https://api.intra.42.fr/oauth/authorize",
				TokenURL: "https://api.intra.42.fr/oauth/token",
			},
		}
	} else {
		fortyTwoOauthConfig = &oauth2.Config {
			RedirectURL: "https://play-stupid.games:30443/api/auth/42/callback",
			ClientID: "u-s4t2ud-a03d8fc82a14a0f36fb4c5e26b33b5414ad93d52a73918090f17c8aa4a9f6364",
			ClientSecret: "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",
			Scopes: []string{"public"},
			Endpoint:	oauth2.Endpoint {
				AuthURL: "https://api.intra.42.fr/oauth/authorize",
				TokenURL: "https://api.intra.42.fr/oauth/token",
			},
		}
	}
	
	//runtime.NumGoroutine()

	serverVars := NewServerStructure()

	defer serverVars.db.GetPool().Close()
	
	gin.SetMode(gin.ReleaseMode)
	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies 

	serverVars.router.Static("/assets", "./static/assets")
	serverVars.router.StaticFile("/favicon.ico", "./static/favicon.ico")
	serverVars.router.NoRoute(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.File("./static/index.html")
	})

	serverVars.router.GET("/api/rooms/:code", func(c *gin.Context) {
		findRoom(c, serverVars)
	})
		serverVars.router.GET("/api/ai-rooms/:code", func(c *gin.Context) {
		findRoom(c, serverVars)
	})
	serverVars.router.GET("/ping", pong)
	serverVars.router.GET("/health", health)
	serverVars.router.GET("/api/config", vaultstatus)
	serverVars.router.GET("/ws", func (c *gin.Context){
		handleWebsocket(c, serverVars)
	})
	serverVars.router.POST("/api/auth/player", func (c *gin.Context){
		handleGuestAuth(c, &serverVars.db)
	})


	// NEW LOGIN CODE
	serverVars.router.GET("/api/auth/42/url", func (c *gin.Context){
		fmt.Println("ATTEMPTING TO GET LOGIN/42/URL FROM ROUTER")
		url := fortyTwoOauthConfig.AuthCodeURL(oauthStateString)
		c.JSON(http.StatusOK, gin.H{"url": url})
	})

	// CALLBACK FOR OAUTH2 WITH 42API

	serverVars.router.GET("/api/auth/42/callback", func(c *gin.Context){
		fmt.Println("42 CALLBACK URL")
		FortyTwoCallback(c, &serverVars.db)
	})

	serverVars.router.POST("/api/auth/register", func(c *gin.Context){
		handleRegister(c, &serverVars.db)
	})

	serverVars.router.POST("/api/auth/login", func(c *gin.Context){
		handleLogin(c, &serverVars.db)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}


	if (os.Getenv("LOCAL") != "") {
		fmt.Println("HOSTING TRANSCENDENCE LOCALLY...")
		if err := serverVars.router.Run(":" + port); err != nil {
			log.Fatalf("Failed to run server: %v", err)	
		}
	} else {

		tlsContent := addnewlinestotls()
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
		}

		fmt.Println(" ~~ Attempting to boot with mTLS on port ", port, " ~~")

		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to run server over mTLS: %v", err)
		}
	}
}
