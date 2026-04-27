package main

import (
	"fmt"
	"sync"
	"os"
	"os/signal"
	"syscall"
	"context"
	
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"crypto/tls"
	"crypto/x509"
	
	"main.go/metrics"
	// "github.com/prometheus/client_golang/prometheus"
)


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

func startupUserMetrics (dbs *DBSafe) {
	db := dbs.GetPool()

	startupQuery := `SELECT
						COUNT(*),
						COUNT(*) FILTER (WHERE type = 'standard'),
						COUNT(*) FILTER (WHERE type = 'guest'),
						COUNT(*) FILTER (WHERE type = 'api42')
 					FROM users;`

	var total float64
	var standard float64
	var guests float64
	var api float64

	err := db.QueryRow(context.Background(), startupQuery).Scan(&total, &standard, &guests, &api)
	if (err != nil) {
		fmt.Printf("Couldn't get the number of users in the database: %v", err)
		return
	}

	metrics.UserCountTotal.Add(total)
	metrics.UserCountStandard.Add(standard)
	metrics.UserCountGuest.Add(guests)
	metrics.UserCountAPI.Add(api)
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

	var connection_url string

	if (os.Getenv("LOCAL") != "") {
		connection_url = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name)
	} else {
		connection_url = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=verify-full", user, pass, host, port, name)
	}

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
