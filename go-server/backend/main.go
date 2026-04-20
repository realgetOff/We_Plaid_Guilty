package main

import (
	"context"
	"fmt"
	"log"
	"net/http" // UNCOMMENT FOR CI/CD DEPLOYMENT
	"os"
	"os/signal"
	"strings"
	"sync"
	"crypto/tls"
	"crypto/x509"
	// "os/signal"
	// "strings"
	"syscall"
	"github.com/joho/godotenv"
	"main.go/gamemanager"
	"runtime"

	// following two are for lobby generation
	//"math/rand/v2"
	// "sync"

	// "github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promauto"
	// "github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zsais/go-gin-prometheus"
	"github.com/prometheus/client_golang/prometheus"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type DBSafe struct {
	mu sync.RWMutex
	Pool *pgxpool.Pool
}

type serverVarsStruct struct { // the name is temporary
	globalHub *gamemanager.Hub
	ClientHub *ClientHub
	router *gin.Engine
	db DBSafe
	//db *pgxpool.Pool
}

// this does nothing? whats the point?
// it doesnt ensure thread safety or anything, it just slows down getting the pgxpool slightly

func (d *DBSafe) GetPool() (pool *pgxpool.Pool){ 
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Pool
}

func reloadConfig(sdb *DBSafe) {
	
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	for {
		<-c
		oldpool := sdb.Pool
		myMap, _ := godotenv.Read("/vault/secrets/app/config")
		db_host := myMap["DB_HOST"]
		db_port := myMap["DB_PORT"]
		db_user := myMap["DB_USER"]
		db_password := myMap["DB_PASSWORD"]
		db_name := myMap["DB_NAME"]

		connection_url := "postgres://" + db_user + ":" + db_password + "@" + db_host + ":" + db_port + "/" + db_name
		content, err := os.ReadFile("/vault/secrets/tls")
		if err != nil { continue }

		cert, err := tls.X509KeyPair(content, content)
		if err != nil { continue }

		cfg, err := pgxpool.ParseConfig(connection_url)
		if err != nil { continue }

		rootCAs := x509.NewCertPool()
		rootCAs.AppendCertsFromPEM(content)

		cfg.ConnConfig.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs: rootCAs,
			InsecureSkipVerify: false,
			ServerName:			db_host,
		}
		fmt.Println("Attempting to connect to :" + connection_url)

		tmppool, err := pgxpool.NewWithConfig(context.Background(), cfg)		
		if err != nil {
			continue
		}
		sdb.mu.Lock()
		sdb.Pool = tmppool
		sdb.mu.Unlock()
		if oldpool != nil {
			oldpool.Close()
		}

	}
}

func connectToDatabase () (*pgxpool.Pool, error) {

	myMap, err := godotenv.Read("/vault/secrets/app/config")

	var host, port, user, pass, name string

	if err == nil {
		host = myMap["DB_HOST"]
		port = myMap["DB_PORT"]
		user = myMap["DB_USER"]
		pass = myMap["DB_PASSWORD"]
		name = myMap["DB_NAME"]
	} else {
		host = os.Getenv("DB_HOST")
		port = os.Getenv("DB_PORT")
		user = os.Getenv("DB_USER")
		pass = os.Getenv("DB_PASSWORD")
		name = os.Getenv("DB_NAME")
	}

	connection_url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name)

	cfg, err := pgxpool.ParseConfig(connection_url)
	if err != nil { return nil, err }

	content, err := os.ReadFile("/vault/secrets/tls")
	if err == nil {
		cert, err := tls.X509KeyPair(content, content)
		if err == nil {

			rootCAs := x509.NewCertPool()
			rootCAs.AppendCertsFromPEM(content)

			cfg.ConnConfig.TLSConfig = &tls.Config{
				Certificates:		[]tls.Certificate{cert},
				RootCAs:		 	rootCAs,
				InsecureSkipVerify:	false,
				ServerName:			host,
			}
			fmt.Println("TLS configuration applied for first connection")

		}
	}

	fmt.Println("Attempting initial connection to DB...")
	return pgxpool.NewWithConfig(context.Background(), cfg)
}

func NewServerStructure () *serverVarsStruct {
	
	// Try to connect to the database, fatally exit if we can't reach it

	var dbs DBSafe

	db, err := connectToDatabase()
	
	dbs.Pool = db
	go reloadConfig(&dbs)
	if err != nil {
		log.Fatalf("Couldn't connect to the PostgreSQL database: %v", err)
	}
	defer db.Close()

	// globalHub := &gamemanager.Hub{
		// Rooms: make(map[string]gamemanager.GameRoom),
	// }

	dbPool, err := connectToDatabase()
	dbs.Pool = dbPool
	go reloadConfig(&dbs)
	if err != nil {
		log.Fatalf("Couldn't connect to the PostgreSQL database: %v", err)
	}
	hub := &gamemanager.Hub{
		Rooms: make(map[string]gamemanager.GameRoom),
	}
	r := gin.Default();

	// PROMETHEUS START
	prometheus.MustRegister(activeWebsockets)
	
	gin_prom := ginprometheus.NewPrometheus("app")
	gin_prom.Use(r)

	// PROMETHEUS END

	chub := &ClientHub{
		Clients:	make(map[string]*Client),
		Db:			dbPool,
	}

	return &serverVarsStruct{
		globalHub:		hub,
		router:			r,
		db:				dbs,
		ClientHub:		chub,
	}
}

func addnewlinestotls() []byte {
	content, err := os.ReadFile("/vault/secrets/tls")
	if err != nil {
		return nil
	}
	delimiter := "-----END CERTIFICATE-----"
	replacement := delimiter + "\n"
	return []byte(strings.ReplaceAll(string(content), delimiter, replacement))
}

// https://api.intra.42.fr/apidoc/guides/web_application_flow#exchange-your-code-for-an-access-token
// https://pkg.go.dev/golang.org/x/oauth2#Endpoint

var (
	fortyTwoOauthConfig = &oauth2.Config {
		RedirectURL: "http://localhost:8080/api/auth/42/callback",
		ClientID: "u-s4t2ud-a03d8fc82a14a0f36fb4c5e26b33b5414ad93d52a73918090f17c8aa4a9f6364",
		ClientSecret: "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",
		Scopes: []string{"public"},
		Endpoint:	oauth2.Endpoint {
			AuthURL: "https://api.intra.42.fr/oauth/authorize",
			TokenURL: "https://api.intra.42.fr/oauth/token",
		},
	}

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback", // change this to whatever it should be IN HTTPS
		ClientID:     "578705584934-5ea9rigvgndb3u1nfm22krmhra3mp9hl.apps.googleusercontent.com", // google client ID from console.cloud.google.com
		ClientSecret: "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", // ditto for client secret
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", 
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	// this should be turned into a randomly generated string
	oauthStateString = "pseudo-random-state"
)

var ( // PROMETHEUS METRICS	
	activeWebsockets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_websockets",
		Help: "The current number of open / active websocket connections.",
	})
)

func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")

	runtime.NumGoroutine()

	serverVars := NewServerStructure()

	defer serverVars.db.GetPool().Close()

	// if err := loadSecretsFromVault(); err != nil {
		// log.Fatalf("Failed to load secrets from Vault: %v", err)
	// }
	// Gin router with default "middleware"
	
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

	serverVars.router.GET("/login/google", func (c *gin.Context){
		url := googleOauthConfig.AuthCodeURL(oauthStateString)
		c.Redirect(http.StatusTemporaryRedirect, url)
	})

	// NEW LOGIN CODE
	serverVars.router.GET("/api/auth/42/url", func (c *gin.Context){
		fmt.Println("ATTEMPTING TO GET LOGIN/42/URL FROM ROUTER")
		url := fortyTwoOauthConfig.AuthCodeURL(oauthStateString)
		c.JSON(http.StatusOK, gin.H{"url": url})
	})

	serverVars.router.GET("/api/auth/42/callback", func(c *gin.Context){
		fmt.Println("42 CALLBACK URL")
		FortyTwoCallback(c, &serverVars.db)
	})

	// need callback functions but im lost at the moment
	// serverVars.router.GET("/auth/42/callback", ...)
	// serverVars.router.GET("/auth/google/callback", ...) 

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// -- OLD ROUTER CODE -- //

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
