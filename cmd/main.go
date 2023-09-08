package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/golden-vcr/tapes/internal/bucket"
	"github.com/golden-vcr/tapes/internal/config"
	"github.com/golden-vcr/tapes/internal/server"
	"github.com/golden-vcr/tapes/internal/sheets"
)

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("error loading .env file: %v", err)
	}
	vars, err := config.LoadVars()
	if err != nil {
		log.Fatalf("error parsing config vars: %v", err)
	}

	sheetsClient, err := sheets.NewClient(context.Background(), vars.SheetsApiKey, vars.SpreadsheetId, time.Hour)
	if err != nil {
		log.Fatalf("error initializing sheets API client: %v", err)
	}
	bucketClient, err := bucket.NewClient(context.Background(), vars.SpacesAccessKeyId, vars.SpacesSecretKey, vars.SpacesEndpointUrl, vars.SpacesRegionName, vars.SpacesBucketName, time.Hour)
	if err != nil {
		log.Fatalf("error initializing S3 bucket API client: %v", err)
	}
	srv := server.New(sheetsClient, bucketClient)

	addr := fmt.Sprintf("%s:%d", vars.BindAddr, vars.ListenPort)
	fmt.Printf("Listening on %s...\n", addr)
	err = http.ListenAndServe(addr, srv)
	if err == http.ErrServerClosed {
		fmt.Printf("Server closed.\n")
	} else {
		log.Fatalf("error running server: %v", err)
	}
}
