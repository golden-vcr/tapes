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
	fmt.Printf("Listing tapes in the Golden VCR Inventory spreadsheet (%s)...\n", config.SpreadsheetId)
	sheetsClient := sheets.NewClient(config.SheetsApiKey, config.SpreadsheetId)
	tapes, warnings, err := sheets.ListTapes(ctx, sheetsClient)
	if err != nil {
		log.Fatalf("error listing tapes from spreadsheet: %v", err)
	}
	for _, warning := range warnings {
		fmt.Printf("[WARNING] At row %d: %s\n", warning.RowNumber, warning.Message)
	}
	fmt.Printf("Got %d tapes:\n", len(tapes))
	for _, tape := range tapes {
		fmt.Printf("- %3d | %4d | %3d | %s\n", tape.Id, tape.Year, tape.Runtime, tape.Title)
	}

	// Initialize an S3 client so we can get image URLs and metadata from our Spaces
	// bucket
	bucketClient, err := bucket.NewClient(ctx, config.SpacesAccessKeyId, config.SpacesSecretKey, config.SpacesEndpointUrl, config.SpacesRegionName, config.SpacesBucketName, time.Hour)
	if err != nil {
		log.Fatalf("error initializing S3 bucket API client: %v", err)
	}
	imageDataByTapeId := bucketClient.GetImageData(ctx)

	// Iterate over all tapes in the spreadsheet
	fmt.Printf("Syncing data for up to %d tapes...\n", len(tapes))
	for _, tape := range tapes {
		// Don't sync a tape unless it has at least one image in the bucket
		images, ok := imageDataByTapeId[tape.Id]
		if !ok || len(images) == 0 {
			fmt.Printf("WARNING: Skipping sync for tape %d; it has no images\n", tape.Id)
			continue
		}

		// Store year and runtime as NULL if not specified
		yearValue := sql.NullInt32{Valid: false}
		if tape.Year > 0 {
			yearValue.Valid = true
			yearValue.Int32 = int32(tape.Year)
		}
		runtimeValue := sql.NullInt32{Valid: false}
		if tape.Runtime > 0 {
			runtimeValue.Valid = true
			runtimeValue.Int32 = int32(tape.Runtime)
		}

		// Start a transaction so that we only commit a tape with all available images
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Fatalf("failed to begin database transaction for tape %d", tape.Id)
		}
		defer tx.Rollback()
		q := queries.New(tx)

		// Upsert into the tape table to register our tape with its latest details
		if err := q.SyncTape(ctx, queries.SyncTapeParams{
			ID:      int32(tape.Id),
			Title:   tape.Title,
			Year:    yearValue,
			Runtime: runtimeValue,
		}); err != nil {
			log.Fatalf("failed to sync tape %d: %v", tape.Id, err)
		}

		// Get the metadata for all images associated with this tape, and register each
		// of those images
		for _, image := range images {
			// Parse the image index from the filename, following naming conventions
			parsed := bucket.ParseKey(image.Filename)
			if parsed == nil || parsed.TapeId != tape.Id || parsed.IsThumbnail {
				log.Fatalf("unable to sync image for tape %d: unexpected image data %+v", tape.Id, parsed)
			}

			// Upsert into the image table to register the latest image metadata
			if err := q.SyncImage(ctx, queries.SyncImageParams{
				TapeID:  int32(tape.Id),
				Index:   int32(parsed.ImageIndex),
				Color:   image.Color,
				Width:   int32(image.Width),
				Height:  int32(image.Height),
				Rotated: image.Rotated,
			}); err != nil {
				log.Fatalf("failed to sync image %d for tape %d: %v", parsed.ImageIndex, tape.Id, err)
			}
		}

		// Commit the transaction; we've finished this tape
		fmt.Printf("tape %d: synced with %d images.\n", tape.Id, len(images))
		if err := tx.Commit(); err != nil {
			log.Fatalf("failed to commit database transaction for tape %d", tape.Id)
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
