# tapes

The **tapes**  API is responsible for encapsulating information about the tapes that
are available in the Golden VCR library.

The source of truth for tapes is the [**Golden VCR Inventory** spreadsheet](https://docs.google.com/spreadsheets/d/1cR9Lbw9_VGQcEn8eGD2b5MwGRGzKugKZ9PVFkrqmA7k/edit#gid=0)
on Google Sheets. This application uses the Google Sheets API in order to access the
data contained in that spreadsheet.

The tapes API is also responsible for knowing which images are available for any given
tape, and where they can be found. Tape images are stored in an S3-compatible bucket in
DigitalOcean Spaces: this application uses the AWS SDK to query that bucket for a
listing of image files.

## Prerequisites

Install [Go 1.21](https://go.dev/doc/install). If successful, you should be able to run:

```
> go version
go version go1.21.0 windows/amd64
```

## Running

Create a file in the root of this repo called `.env` that contains the environment
variables required in [`main.go`](./cmd/server/main.go). If you have the
[`terraform`](https://github.com/golden-vcr/terraform) repo cloned alongside this one,
simply open a shell there and run:

- `terraform output -raw sheets_api_env > ../tapes/.env && terraform output -raw images_s3_env >> ../tapes/.env`

Once your `.env` file is populated, you should be able to build and run the server:

- `go run cmd/server/main.go`

If successful, you should be able to run `curl http://localhost:5000/` and get a
JSON array containing tape data fetched from the configured spreadsheet.
