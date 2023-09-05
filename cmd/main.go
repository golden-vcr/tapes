package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/golden-vcr/tapes/internal/config"
	"github.com/golden-vcr/tapes/internal/sheets"
)

func handleTapes(client *sheets.Client, res http.ResponseWriter, req *http.Request) {
	rows, err := client.ListTapes(req.Context())
	if err == nil {
		res.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(res).Encode(rows)
	}
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}
}

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("error loading .env file: %v", err)
	}

	vars, err := config.LoadVars()
	if err != nil {
		log.Fatalf("error parsing config vars: %v", err)
	}

	client, err := sheets.NewClient(context.Background(), vars.SheetsApiKey, vars.SpreadsheetId, time.Hour)
	if err != nil {
		log.Fatalf("error initializing sheets API client: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/tapes", func(res http.ResponseWriter, req *http.Request) {
		handleTapes(client, res, req)
	})

	addr := fmt.Sprintf("%s:%d", vars.BindAddr, vars.ListenPort)
	fmt.Printf("Listening on %s...\n", addr)
	err = http.ListenAndServe(addr, mux)
	if err == http.ErrServerClosed {
		fmt.Printf("Server cloesd.\n")
	} else {
		log.Fatalf("error running server: %v", err)
	}
}
