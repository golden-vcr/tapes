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

	"github.com/golden-vcr/tapes/internal/bucket"
	"github.com/golden-vcr/tapes/internal/config"
	"github.com/golden-vcr/tapes/internal/sheets"
)

type TapeListing struct {
	Tapes        []TapeListingItem `json:"tapes"`
	ImageHostUrl string            `json:"imageHostUrl"`
}

type TapeListingItem struct {
	Id                     int      `json:"id"`
	Title                  string   `json:"title"`
	Year                   int      `json:"year"`
	RuntimeMinutes         int      `json:"runtime"`
	ThumbnailImageFilename string   `json:"thumbnailImageFilename"`
	ImageFilenames         []string `json:"imageFilenames"`
}

func handleGetTapeListing(sheetsClient *sheets.Client, bucketClient *bucket.Client, res http.ResponseWriter, req *http.Request) {
	rows, err := sheetsClient.ListTapes(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	imageCounts, err := bucketClient.GetImageCounts(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]TapeListingItem, 0, len(rows))
	for _, row := range rows {
		numImages, ok := imageCounts[row.ID]
		if !ok || numImages <= 0 {
			continue
		}
		imageFilenames := make([]string, 0, numImages)
		for imageIndex := 0; imageIndex < numImages; imageIndex++ {
			imageFilenames = append(imageFilenames, bucket.GetImageKey(row.ID, imageIndex))
		}
		items = append(items, TapeListingItem{
			Id:                     row.ID,
			Title:                  row.Title,
			Year:                   row.Year,
			RuntimeMinutes:         row.RuntimeMin,
			ThumbnailImageFilename: bucket.GetThumbnailKey(row.ID),
			ImageFilenames:         imageFilenames,
		})
	}

	result := TapeListing{
		Tapes:        items,
		ImageHostUrl: bucketClient.GetImageHostURL(),
	}
	if err := json.NewEncoder(res).Encode(result); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
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

	sheetsClient, err := sheets.NewClient(context.Background(), vars.SheetsApiKey, vars.SpreadsheetId, time.Hour)
	if err != nil {
		log.Fatalf("error initializing sheets API client: %v", err)
	}

	bucketClient, err := bucket.NewClient(context.Background(), vars.SpacesAccessKeyId, vars.SpacesSecretKey, vars.SpacesEndpointUrl, vars.SpacesRegionName, vars.SpacesBucketName, time.Hour)
	if err != nil {
		log.Fatalf("error initializing S3 bucket API client: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/tapes", func(res http.ResponseWriter, req *http.Request) {
		handleGetTapeListing(sheetsClient, bucketClient, res, req)
	})

	addr := fmt.Sprintf("%s:%d", vars.BindAddr, vars.ListenPort)
	fmt.Printf("Listening on %s...\n", addr)
	err = http.ListenAndServe(addr, mux)
	if err == http.ErrServerClosed {
		fmt.Printf("Server closed.\n")
	} else {
		log.Fatalf("error running server: %v", err)
	}
}
