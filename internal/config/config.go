package config

import (
	"github.com/codingconcepts/env"
)

type Vars struct {
	BindAddr   string `env:"BIND_ADDR"`
	ListenPort uint16 `env:"LISTEN_PORT" default:"5000"`

	SheetsApiKey  string `env:"SHEETS_API_KEY" required:"true"`
	SpreadsheetId string `env:"SPREADSHEET_ID" required:"true"`

	SpacesBucketName  string `env:"SPACES_BUCKET_NAME" required:"true"`
	SpacesRegionName  string `env:"SPACES_REGION_NAME" required:"true"`
	SpacesEndpointUrl string `env:"SPACES_ENDPOINT_URL" required:"true"`
	SpacesAccessKeyId string `env:"SPACES_ACCESS_KEY_ID" required:"true"`
	SpacesSecretKey   string `env:"SPACES_SECRET_KEY" required:"true"`
}

func LoadVars() (*Vars, error) {
	vars := Vars{}
	if err := env.Set(&vars); err != nil {
		return nil, err
	}
	return &vars, nil
}
