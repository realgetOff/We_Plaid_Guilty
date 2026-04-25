package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" // Driver Postgres
)

func main() {
	// 1. Récupération de la config (via env)
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", 
		dbHost, dbUser, dbPass, dbName) // NOTE for ssl certif with infra sslmode=verify-full

	// 2. Test de connexion DB
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 3. Simple API pour vérifier que ça répond
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		err := db.Ping()
		if err != nil {
			http.Error(w, "DB down", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, "Pong! Backend Go et Postgres sont copains.")
	})

	log.Println("Serveur lancé sur le port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
