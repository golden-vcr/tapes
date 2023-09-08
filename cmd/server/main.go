package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/codingconcepts/env"
	"github.com/joho/godotenv"

	"github.com/golden-vcr/tapes/internal/bucket"
	"github.com/golden-vcr/tapes/internal/server"
	"github.com/golden-vcr/tapes/internal/sheets"
)

type Config struct {
	BindAddr   string `env:"BIND_ADDR"`
	ListenPort uint16 `env:"LISTEN_PORT" default:"5000"`

	SheetsApiKey  string `env:"SHEETS_API_KEY" required:"true"`
	SpreadsheetId string `env:"SPREADSHEET_ID" required:"true"`

	SpacesBucketName  string `env:"SPACES_BUCKET_NAME" required:"true"`
	SpacesRegionName  string `env:"SPACES_REGION_NAME" required:"true"`
	SpacesEndpointUrl string `env:"SPACES_ENDPOINT_URL" required:"true"`
	SpacesAccessKeyId string `env:"SPACES_ACCESS_KEY_ID" required:"true"`
	SpacesSecretKey   string `env:"SPACES_SECRET_KEY" required:"true"`
}

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("error loading .env file: %v", err)
	}
	config := Config{}
	if err := env.Set(&config); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	sheetsClient, err := sheets.NewClient(context.Background(), config.SheetsApiKey, config.SpreadsheetId, time.Hour)
	if err != nil {
		log.Fatalf("error initializing sheets API client: %v", err)
	}
	bucketClient, err := bucket.NewClient(context.Background(), config.SpacesAccessKeyId, config.SpacesSecretKey, config.SpacesEndpointUrl, config.SpacesRegionName, config.SpacesBucketName, time.Hour)
	if err != nil {
		log.Fatalf("error initializing S3 bucket API client: %v", err)
	}
	srv := server.New(sheetsClient, bucketClient)

	addr := fmt.Sprintf("%s:%d", config.BindAddr, config.ListenPort)
	fmt.Printf("Listening on %s...\n", addr)
	err = http.ListenAndServe(addr, srv)
	if err == http.ErrServerClosed {
		fmt.Printf("Server closed.\n")
	} else {
		log.Fatalf("error running server: %v", err)
	}
}
