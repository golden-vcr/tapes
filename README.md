# tapes

The **tapes** API is responsible for encapsulating information about the tapes that are
available in the Golden VCR library.

The source of truth for tapes is the [**Golden VCR Inventory** spreadsheet](https://docs.google.com/spreadsheets/d/1cR9Lbw9_VGQcEn8eGD2b5MwGRGzKugKZ9PVFkrqmA7k/edit#gid=0)
on Google Sheets. This application uses the Google Sheets API in order to access the
data contained in that spreadsheet.

The tapes API is also responsible for knowing which images are available for any given
tape, and where they can be found. Tape images are stored in an S3-compatible bucket in
DigitalOcean Spaces: this application uses the AWS SDK to query that bucket for a
listing of image files.

## Development Guide

On a Linux or WSL system:

1. Install [Go 1.21](https://go.dev/doc/install)
2. Clone the [**terraform**](https://github.com/golden-vcr/terraform) repo alongside
   this one, and from the root of that repo:
    - Ensure that the module is initialized (via `terraform init`)
    - Ensure that valid terraform state is present
    - Run `terraform output -raw env_tapes_local > ../tapes/.env` to populate an `.env`
      file.
    - Run [`./local-db.sh up`](https://github.com/golden-vcr/terraform/blob/main/local-db.sh)
      to ensure that a Postgres server is running locally (requires
      [Docker](https://docs.docker.com/engine/install/)).
3. Ensure that the [**auth**](https://github.com/golden-vcr/auth?tab=readme-ov-file#development-guide)
   server is running locally.
4. From the root of this repository:
    - Run [`./db-migrate.sh`](./db-migrate.sh) to apply database migrations.
    - Run [`go run cmd/sync/main.go`](./cmd/sync/main.go) to sync tape data from the
      spreadsheet to the local database.
    - Run [`go run cmd/server/main.go`](./cmd/server/main.go) to start up the server.

Once done, the tapes server will be running at http://localhost:5000.

### Generating database queries

If you modify the SQL code in [`db/queries`](./db/queries/), you'll need to generate
new Go code to [`gen/queries`](./gen/queries/). To do so, simply run:

- `./db-generate-queries.sh`
