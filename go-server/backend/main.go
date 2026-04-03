package main

import (
	//"encoding/json"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"crypto/tls"
	"crypto/x509"
	// "os/signal"
	"strings"
	"syscall"
	"github.com/joho/godotenv"
	"main.go/gamemanager"

	// following two are for lobby generation
	//"math/rand/v2"
	// "sync"

	"github.com/gin-gonic/gin"
	//"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
The message structure contains the json information to be sent / received by the websocket for room generation
type: state before / after generation of the room code
code: room code
omitempty: omits empty strings, lowering network traffic

*/

type DBSafe struct {
	mu sync.RWMutex
	Pool *pgxpool.Pool
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
		myMap, _ := godotenv.Read("/vault/data/app/config")
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

var globalHub *gamemanager.Hub

func connectToDatabase () (*pgxpool.Pool, error) {

	myMap, err := godotenv.Read("/vault/data/app/config")

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
				Certificates: []tls.Certificate{cert},
				RootCAs:      rootCAs,
				InsecureSkipVerify: false,
			}
			fmt.Println("TLS configuration applied for first connection")

		}
	}

	fmt.Println("Attempting initial connection to DB...")
	return pgxpool.NewWithConfig(context.Background(), cfg)
}


func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")

	var dbs DBSafe
	db, err := connectToDatabase()
	
	dbs.Pool = db
	go reloadConfig(&dbs)
	if err != nil {
		log.Fatalf("Couldn't connect to the PostgreSQL database: %v", err)
	}
	defer db.Close()

	globalHub = &gamemanager.Hub{
		Rooms: make(map[string]gamemanager.GameRoom),
	}


	// if err := loadSecretsFromVault(); err != nil {
	// 	log.Fatalf("Failed to load secrets from Vault: %v", err)
	// }
	// Gin router with default "middleware"
	router := gin.Default();
	// gin.SetMode(gin.ReleaseMode)
	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies 

	router.Static("/assets", "./static/assets")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	router.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})


	router.GET("/api/ai-rooms/:code", func(c *gin.Context) {
    code := strings.ToUpper(c.Param("code"))
    room, err := globalHub.GetRoom(code)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "ai room not found"})
        return
    }
	base := room.GetBase()
    c.JSON(http.StatusOK, gin.H{
        "code":    base.ID,
        "status":  base.Status,
        "players": len(base.Players),
    })
})
	router.GET("/api/rooms/:code", func(c *gin.Context) {
		code := strings.ToUpper(c.Param("code"))
		room, err := globalHub.GetRoom(code)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		
		base := room.GetBase()
		c.JSON(http.StatusOK, gin.H{
			"code":    base.ID,
			"status":  base.Status,
			"players": len(base.Players),
		})
	})
	router.GET("/ping", pong)
	router.GET("/health", health)
	router.GET("/api/config", vaultstatus)
	router.GET("/ws", func (c *gin.Context){
		handleWebsocket(c, &dbs, globalHub)
	})
	router.POST("/api/auth/player", func (c *gin.Context){
		handleGuestAuth(c, &dbs)
	})

	// get the port defined in the environment variables, if theres fuckall, 8080

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
		
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)	
	}
}
