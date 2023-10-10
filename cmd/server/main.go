package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/codingconcepts/env"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/golden-vcr/server-common/db"
	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/golden-vcr/tapes/internal/catalog"
)

type Config struct {
	BindAddr   string `env:"BIND_ADDR"`
	ListenPort uint16 `env:"LISTEN_PORT" default:"5000"`

	SpacesBucketName     string `env:"SPACES_BUCKET_NAME" required:"true"`
	SpacesEndpointOrigin string `env:"SPACES_ENDPOINT_URL" required:"true"`

	DatabaseHost     string `env:"PGHOST" required:"true"`
	DatabasePort     int    `env:"PGPORT" required:"true"`
	DatabaseName     string `env:"PGDATABASE" required:"true"`
	DatabaseUser     string `env:"PGUSER" required:"true"`
	DatabasePassword string `env:"PGPASSWORD" required:"true"`
	DatabaseSslMode  string `env:"PGSSLMODE"`
}

func main() {
	// Parse config from environment variables
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("error loading .env file: %v", err)
	}
	config := Config{}
	if err := env.Set(&config); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Shut down cleanly on signal
	ctx, close := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer close()

	// Configure our database connection and initialize a Queries struct, so we can read
	// from the 'tapes' schema in response to HTTP requests
	connectionString := db.FormatConnectionString(
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
	q := queries.New(db)

	// Start setting up our HTTP handlers, using gorilla/mux for routing
	r := mux.NewRouter()

	// Clients can hit GET /catalog to retrieve information about tapes in the Golden
	// VCR Library
	{
		imageHostUrl := fmt.Sprintf("https://%s.%s", config.SpacesBucketName, config.SpacesEndpointOrigin)
		catalogServer := catalog.NewServer(q, imageHostUrl)
		catalogServer.RegisterRoutes(r.PathPrefix("/catalog").Subrouter())
	}

	addr := fmt.Sprintf("%s:%d", config.BindAddr, config.ListenPort)
	server := &http.Server{Addr: addr, Handler: r}

	// Handle incoming HTTP connections until our top-level context is canceled, at
	// which point shut down cleanly
	fmt.Printf("Listening on %s...\n", addr)
	var wg errgroup.Group
	wg.Go(server.ListenAndServe)

	select {
	case <-ctx.Done():
		fmt.Printf("Received signal; closing server...\n")
		server.Shutdown(context.Background())
	}

	err = wg.Wait()
	if err == http.ErrServerClosed {
		fmt.Printf("Server closed.\n")
	} else {
		log.Fatalf("error running server: %v", err)
	}
}
