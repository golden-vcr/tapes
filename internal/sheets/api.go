package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type valuesResult struct {
	Range          string     `json:"range"`
	MajorDimension string     `json:"majorDimension"`
	Values         [][]string `json:"values"`
}

type errorResult struct {
	Error errorData `json:"error"`
}

type errorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func getSheetValues(ctx context.Context, sheetsApiKey string, spreadsheetId string, sheetName string) (*valuesResult, error) {
	// Build a request to the Google Sheets API to get the full contents of our desired sheet
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s", spreadsheetId, sheetName)
	fmt.Printf("> GET %s\n", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-goog-api-key", sheetsApiKey)

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
	var result valuesResult
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Sheets API response: %w", err)
	}
	return &result, nil
}

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
