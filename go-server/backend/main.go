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

	// following two are for lobby generation
	//"math/rand/v2"
	// "sync"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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

func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// -- OLD ROUTER CODE -- //

	// if err := serverVars.router.Run(":" + port); err != nil {
		// log.Fatalf("Failed to run server: %v", err)	
	// }

	// UNCOMMENT NET/HTTP BEFORE DEPLOYING VIA CI/CD

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
