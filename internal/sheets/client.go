package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SheetName is the name given to the sheet (i.e. tab) within the Golden VCR Inventory
// spreadsheet that should be parsed to obtain the current list of tapes
const SheetName = "Tapes"

// GetValuesResult is the payload returned by the Google Sheets API from
// GET /v4/spreadsheets/:spreadsheetId/values/:sheetName
type GetValuesResult struct {
	Range          string     `json:"range"`
	MajorDimension string     `json:"majorDimension"`
	Values         [][]string `json:"values"`
}

// Client allows the contents of a single sheet in a single spreadsheet to be fetched
// from the Google Sheets API
type Client interface {
	GetValues(ctx context.Context) (*GetValuesResult, error)
}

// NewClient returns a Client that will fetch data from an actual spreadsheet in Google
// sheets
func NewClient(sheetsApiKey string, spreadsheetId string) Client {
	return &client{
		sheetsApiKey:  sheetsApiKey,
		spreadsheetId: spreadsheetId,
		sheetName:     SheetName,
	}
}

// client implementats sheets.Client, configured with a Google API key and a
// spreadsheet ID and sheet name to pull values from
type client struct {
	sheetsApiKey  string
	spreadsheetId string
	sheetName     string
}

// errorResult is the payload that the Sheets API returns to provide more details about
// an error
type errorResult struct {
	Error errorData `json:"error"`
}

// errorData encodes the details of an error encountered during a Sheets API call
type errorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// getSheetValues requests the raw values from a sheet in a Google Sheets spreadhsset,
// authorized with the given API key
func (c *client) GetValues(ctx context.Context) (*GetValuesResult, error) {
	// Build a request to the Google Sheets API to get the full contents of our desired sheet
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s", c.spreadsheetId, c.sheetName)
	fmt.Printf("> GET %s\n", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-goog-api-key", c.sheetsApiKey)

	// Make the request and verify that we got an OK result
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	fmt.Printf("< %d\n", res.StatusCode)
	if err := handleRequestError(res); err != nil {
		return nil, err
	}

	// We have a 200 response from the Sheets API: parse it from JSON
	contentType := res.Header.Get("content-type")
	if !strings.HasPrefix(contentType, "application/json") {
		return nil, fmt.Errorf("expected a response with content-type 'application/json'; got '%s'", contentType)
	}
	var result GetValuesResult
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Sheets API response: %w", err)
	}
	return &result, nil
}

// handleRequestError returns nil if the provided response is OK; otherwise it returns
// an error populated with the error details returned by the Sheets API
func handleRequestError(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	data := readErrorData(res)
	if data != nil {
		return fmt.Errorf("got error %d (%s): %s", data.Code, data.Status, data.Message)
	}
	return fmt.Errorf("got status %d from Sheets API call", res.StatusCode)
}

func readErrorData(res *http.Response) *errorData {
	if !strings.HasPrefix(res.Header.Get("content-type"), "application/json") {
		return nil
	}
	var result errorResult
	err := json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil
	}
	return &result.Error
}
