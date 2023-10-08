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

	"github.com/codingconcepts/env"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/golden-vcr/tapes/internal/sheets"
	"github.com/golden-vcr/tapes/internal/storage"
)

type Config struct {
	SheetsApiKey  string `env:"SHEETS_API_KEY" required:"true"`
	SpreadsheetId string `env:"SPREADSHEET_ID" required:"true"`

	SpacesBucketName     string `env:"SPACES_BUCKET_NAME" required:"true"`
	SpacesRegionName     string `env:"SPACES_REGION_NAME" required:"true"`
	SpacesEndpointOrigin string `env:"SPACES_ENDPOINT_URL" required:"true"`
	SpacesAccessKeyId    string `env:"SPACES_ACCESS_KEY_ID" required:"true"`
	SpacesSecretKey      string `env:"SPACES_SECRET_KEY" required:"true"`

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
	tapes, sheetWarnings, err := sheets.ListTapes(ctx, sheetsClient)
	if err != nil {
		log.Fatalf("error listing tapes from spreadsheet: %v", err)
	}
	fmt.Printf("Got %d tapes:\n", len(tapes))
	for _, tape := range tapes {
		fmt.Printf("- %3d | %4d | %3d | %s\n", tape.Id, tape.Year, tape.Runtime, tape.Title)
	}

	// Initialize an S3 client so we can get image URLs and metadata from our Spaces
	// bucket
	fmt.Printf("Retrieving image filenames and metadata from storage bucket (%s)...\n", config.SpacesBucketName)
	storageClient, err := storage.NewClient(
		config.SpacesAccessKeyId,
		config.SpacesSecretKey,
		config.SpacesEndpointOrigin,
		config.SpacesRegionName,
		config.SpacesBucketName,
	)
	if err != nil {
		log.Fatalf("error initializing client for S3-compatible storage: %v", err)
	}
	images, imageWarnings, err := storage.ListImages(ctx, storageClient)
	if err != nil {
		log.Fatalf("error retrieving image data from storage bucket: %v", err)
	}
	fmt.Printf("Got %d images:\n", len(images))
	for _, image := range images {
		summary := fmt.Sprintf("%3d | %-9s | %s", image.TapeId, image.Type, image.Filename)
		if image.Type == storage.ImageTypeGallery {
			index := image.GalleryData.Index
			md := image.GalleryData.Metadata
			flag := ""
			if md.Rotated {
				flag = "rotated"
			}
			fmt.Printf("- %s | %d | %d x %d | %s | %s\n", summary, index, md.Width, md.Height, md.Color, flag)
		} else {
			fmt.Printf("- %s\n", summary)
		}
	}

	// We don't actually record anything in the database for thumbnail images: we just
	// require that a tape have a thumbnail image before we record that the tape exists,
	// so we can assume that every tape has a thumbnail image at %04d_thumb.jpg. Collect
	// all of the gallery images that we need to record for each tape.
	galleryImagesByTapeId := make(map[int][]*storage.Image)
	for _, image := range images {
		galleryImagesByTapeId[image.TapeId] = append(galleryImagesByTapeId[image.TapeId], &image)
	}

	// Collect a list of all warnings, line-by-line as strings, so we can present a
	// summary when finished syncing
	warningLines := make([]string, 0, len(sheetWarnings)+len(imageWarnings))
	for _, warning := range sheetWarnings {
		warningLines = append(warningLines, fmt.Sprintf("Spreadsheet row %d: %s", warning.RowNumber, warning.Message))
	}
	for _, warning := range imageWarnings {
		warningLines = append(warningLines, fmt.Sprintf("Image file %s: %s", warning.Filename, warning.Message))
	}

	// Iterate over all tapes in the spreadsheet
	fmt.Printf("Syncing tape and image data to the tapes database...\n")
	numTapesSynced := 0
	for _, tape := range tapes {
		// Don't sync a tape unless it has at least one gallery image stored
		galleryImages, ok := galleryImagesByTapeId[tape.Id]
		if !ok || len(galleryImages) == 0 {
			warningLines = append(warningLines, fmt.Sprintf("Tape %d has no image files; ignoring it.", tape.Id))
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
		for _, image := range galleryImages {
			// Upsert into the image table to register the latest image metadata
			if err := q.SyncImage(ctx, queries.SyncImageParams{
				TapeID:  int32(tape.Id),
				Index:   int32(image.GalleryData.Index),
				Color:   string(image.GalleryData.Metadata.Color),
				Width:   int32(image.GalleryData.Metadata.Width),
				Height:  int32(image.GalleryData.Metadata.Height),
				Rotated: image.GalleryData.Metadata.Rotated,
			}); err != nil {
				log.Fatalf("failed to sync image %d for tape %d: %v", image.GalleryData.Index, tape.Id, err)
			}
		}

		// Commit the transaction; we've finished this tape
		fmt.Printf("tape %d: synced with %d images.\n", tape.Id, len(galleryImages))
		if err := tx.Commit(); err != nil {
			log.Fatalf("failed to commit database transaction for tape %d", tape.Id)
		}
		numTapesSynced++
	}

	fmt.Printf("Synced data for %d tape(s).\n", numTapesSynced)
	if len(warningLines) > 0 {
		fmt.Printf("Encountered %d warning(s):\n", len(warningLines))
		for _, line := range warningLines {
			fmt.Printf("- %s\n", line)
		}
	}
}

func formatConnectionString(host string, port int, dbname string, user string, password string, sslmode string) string {
	urlencodedPassword := url.QueryEscape(password)
	s := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, urlencodedPassword, host, port, dbname)
	if sslmode != "" {
		s += fmt.Sprintf("?sslmode=%s", sslmode)
	}
	return s
}
