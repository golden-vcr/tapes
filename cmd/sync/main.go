package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codingconcepts/env"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/golden-vcr/tapes/internal/bucket"
	"github.com/golden-vcr/tapes/internal/sheets"
)

type Config struct {
	SheetsApiKey  string `env:"SHEETS_API_KEY" required:"true"`
	SpreadsheetId string `env:"SPREADSHEET_ID" required:"true"`

	SpacesBucketName  string `env:"SPACES_BUCKET_NAME" required:"true"`
	SpacesRegionName  string `env:"SPACES_REGION_NAME" required:"true"`
	SpacesEndpointUrl string `env:"SPACES_ENDPOINT_URL" required:"true"`
	SpacesAccessKeyId string `env:"SPACES_ACCESS_KEY_ID" required:"true"`
	SpacesSecretKey   string `env:"SPACES_SECRET_KEY" required:"true"`

	DatabaseHost     string `env:"PGHOST" required:"true"`
	DatabasePort     int    `env:"PGPORT" required:"true"`
	DatabaseName     string `env:"PGDATABASE" required:"true"`
	DatabaseUser     string `env:"PGUSER" required:"true"`
	DatabasePassword string `env:"PGPASSWORD" required:"true"`
	DatabaseSslMode  string `env:"PGSSLMODE"`
}

func main() {
	// Load config from .env
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("error loading .env file: %v", err)
	}
	config := Config{}
	if err := env.Set(&config); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Terminate on SIGINT etc.
	ctx, close := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer close()

	// Connect to the tapes database so we can sync data to it from sheets and S3
	connectionString := formatConnectionString(
		config.DatabaseHost,
		config.DatabasePort,
		config.DatabaseName,
		config.DatabaseUser,
		config.DatabasePassword,
		config.DatabaseSslMode,
	)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	// Initialize a Google sheets API client and get a listing of all tapes with valid
	// rows in the inventory spreadsheet
	sheetsClient, err := sheets.NewClient(context.Background(), config.SheetsApiKey, config.SpreadsheetId, time.Hour)
	if err != nil {
		log.Fatalf("error initializing sheets API client: %v", err)
	}
	rows, err := sheetsClient.ListTapes(ctx)
	if err != nil {
		log.Fatalf("error listing rows from sheets API: %v", err)
	}

	// Initialize an S3 client so we can get image URLs and metadata from our Spaces
	// bucket
	bucketClient, err := bucket.NewClient(context.Background(), config.SpacesAccessKeyId, config.SpacesSecretKey, config.SpacesEndpointUrl, config.SpacesRegionName, config.SpacesBucketName, time.Hour)
	if err != nil {
		log.Fatalf("error initializing S3 bucket API client: %v", err)
	}
	imageDataByTapeId := bucketClient.GetImageData(ctx)

	// Iterate over all tapes in the spreadsheet
	fmt.Printf("Syncing data for up to %d tapes...\n", len(rows))
	for _, row := range rows {
		// Don't sync a tape unless it has at least one image in the bucket
		images, ok := imageDataByTapeId[row.ID]
		if !ok || len(images) == 0 {
			fmt.Printf("WARNING: Skipping sync for tape %d; it has no images\n", row.ID)
			continue
		}

		// Store year and runtime as NULL if not specified
		yearValue := sql.NullInt32{Valid: false}
		if row.Year > 0 {
			yearValue.Valid = true
			yearValue.Int32 = int32(row.Year)
		}
		runtimeValue := sql.NullInt32{Valid: false}
		if row.RuntimeMin > 0 {
			runtimeValue.Valid = true
			runtimeValue.Int32 = int32(row.RuntimeMin)
		}

		// Start a transaction so that we only commit a tape with all available images
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Fatalf("failed to begin database transaction for tape %d", row.ID)
		}
		defer tx.Rollback()
		q := queries.New(tx)

		// Upsert into the tape table to register our tape with its latest details
		if err := q.SyncTape(ctx, queries.SyncTapeParams{
			ID:      int32(row.ID),
			Title:   row.Title,
			Year:    yearValue,
			Runtime: runtimeValue,
		}); err != nil {
			log.Fatalf("failed to sync tape %d: %v", row.ID, err)
		}

		// Get the metadata for all images associated with this tape, and register each
		// of those images
		for _, image := range images {
			// Parse the image index from the filename, following naming conventions
			parsed := bucket.ParseKey(image.Filename)
			if parsed == nil || parsed.TapeId != row.ID || parsed.IsThumbnail {
				log.Fatalf("unable to sync image for tape %d: unexpected image data %+v", row.ID, parsed)
			}

			// Upsert into the image table to register the latest image metadata
			if err := q.SyncImage(ctx, queries.SyncImageParams{
				TapeID:  int32(row.ID),
				Index:   int32(parsed.ImageIndex),
				Color:   image.Color,
				Width:   int32(image.Width),
				Height:  int32(image.Height),
				Rotated: image.Rotated,
			}); err != nil {
				log.Fatalf("failed to sync image %d for tape %d: %v", parsed.ImageIndex, row.ID, err)
			}
		}

		// Commit the transaction; we've finished this tape
		fmt.Printf("tape %d: synced with %d images.\n", row.ID, len(images))
		if err := tx.Commit(); err != nil {
			log.Fatalf("failed to commit database transaction for tape %d", row.ID)
		}
	}
	fmt.Printf("Done.\n")
}

func formatConnectionString(host string, port int, dbname string, user string, password string, sslmode string) string {
	urlencodedPassword := url.QueryEscape(password)
	s := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, urlencodedPassword, host, port, dbname)
	if sslmode != "" {
		s += fmt.Sprintf("?sslmode=%s", sslmode)
	}
	return s
}
