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
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"

	"github.com/golden-vcr/auth"
	"github.com/golden-vcr/server-common/db"
	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/golden-vcr/tapes/internal/catalog"
	"github.com/golden-vcr/tapes/internal/favorites"
	"github.com/golden-vcr/tapes/internal/users"
)

type Config struct {
	BindAddr   string `env:"BIND_ADDR"`
	ListenPort uint16 `env:"LISTEN_PORT" default:"5000"`

	SpacesBucketName     string `env:"SPACES_BUCKET_NAME" required:"true"`
	SpacesEndpointOrigin string `env:"SPACES_ENDPOINT_URL" required:"true"`

	TwitchClientId          string `env:"TWITCH_CLIENT_ID" required:"true"`
	TwitchClientSecret      string `env:"TWITCH_CLIENT_SECRET" required:"true"`
	TwitchExtensionClientId string `env:"TWITCH_EXTENSION_CLIENT_ID" required:"true"`

	AuthURL string `env:"AUTH_URL" default:"http://localhost:5002"`

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

	// We use a simple Twitch API client in order to resolve user-facing display names
	// (from Twitch User IDs) for tapes that were contributed by a specific user
	lookup, err := users.NewLookup(config.TwitchClientId, config.TwitchClientSecret)
	if err != nil {
		log.Fatalf("error initializing user lookup: %v", err)
	}

	// Some requests carry a user authorization token identifying the user, which is
	// required for certain features (e.g. keeping track of users' favorite tapes)
	authClient := auth.NewClient(config.AuthURL)

	// Start setting up our HTTP handlers, using gorilla/mux for routing
	r := mux.NewRouter()

	// Clients can hit GET /catalog to retrieve information about tapes in the Golden
	// VCR Library
	{
		imageHostUrl := fmt.Sprintf("https://%s.%s", config.SpacesBucketName, config.SpacesEndpointOrigin)
		catalogServer := catalog.NewServer(q, lookup, imageHostUrl)
		catalogServer.RegisterRoutes(r.PathPrefix("/catalog").Subrouter())
	}

	// Once logged in, users can hit GET /favorites to get the set of tape IDs that a
	// user has selected as their favorites, and can use PATCH /favorites to change
	// their favorite tape selection
	{
		favoritesServer := favorites.NewServer(q)
		favoritesServer.RegisterRoutes(authClient, r.PathPrefix("/favorites").Subrouter())
	}

	// Inject CORS support, allowing the Twitch-hosted extension to make read-only
	// requests to all tapes API endpoints
	withCors := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://localhost:5180",
			fmt.Sprintf("https://%s.ext-twitch.tv", config.TwitchExtensionClientId),
		},
		AllowedMethods: []string{http.MethodGet},
	})
	addr := fmt.Sprintf("%s:%d", config.BindAddr, config.ListenPort)
	server := &http.Server{Addr: addr, Handler: withCors.Handler(r)}

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
