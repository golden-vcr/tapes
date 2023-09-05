# tapes

The **tapes**  API is responsible for encapsulating information about the tapes that
are available in the Golden VCR library.

The source of truth for tapes is the [**Golden VCR Inventory** spreadsheet](https://docs.google.com/spreadsheets/d/1cR9Lbw9_VGQcEn8eGD2b5MwGRGzKugKZ9PVFkrqmA7k/edit#gid=0)
on Google Sheets. This server uses the Google Sheets API in order to access the data
contained in that spreadsheet.

## Prerequisites

Install [Go 1.21](https://go.dev/doc/install). If successful, you should be able to run:

```
> go version
go version go1.21.0 windows/amd64
```

## Running

Define the environment variables required in [`config.go`](./internal/config/config.go),
then build and run the program:

```
export SHEETS_API_KEY="..."
export SPREADSHEET_ID="1cR9Lbw9_VGQcEn8eGD2b5MwGRGzKugKZ9PVFkrqmA7k"
go run cmd/main.go
```

If successful, you should be able to run `curl http://localhost:5000/tapes` and get a
JSON array containing tape data fetched from the configured spreadsheet.
