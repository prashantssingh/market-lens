package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func main() {
	dataDir := env("DATA_DIR", "./data")
	if err := os.MkdirAll(filepath.Join(dataDir, "snapshots"), 0755); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite", filepath.Join(dataDir, "market-lens.db"))
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	app := newApp(db, dataDir, os.Getenv("ALPHA_VANTAGE_API_KEY"))
	if err := app.migrate(); err != nil {
		log.Fatal(err)
	}

	addr := ":" + env("PORT", "8080")
	log.Printf("Market Lens API listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, withCORS(app.routes())))
}
