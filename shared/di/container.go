package di

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
	dbgen "github.com/bluelock-go/shared/database/generated"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

// Container holds all application dependencies
type Container struct {
	Logger       *shared.CustomLogger
	Config       *config.Config
	StateManager *statemanager.StateManager
	Credentials  []auth.Credential
	DB           *sql.DB
	DBQueries    *dbgen.Queries
	HTTPClient   *http.Client
}

// Ensure Container implements all provider interfaces
var _ LoggerProvider = (*Container)(nil)
var _ StateManagerProvider = (*Container)(nil)
var _ CredentialsProvider = (*Container)(nil)
var _ HTTPClientProvider = (*Container)(nil)
var _ ConfigProvider = (*Container)(nil)
var _ DatabaseProvider = (*Container)(nil)
var _ APIProvider = (*Container)(nil)
var _ ServiceProvider = (*Container)(nil)
var _ FullProvider = (*Container)(nil)

// NewContainer returns a new Container instance with the HTTP client initialized to the default client.
func NewContainer() *Container {
	return &Container{
		HTTPClient: http.DefaultClient,
	}
}

// Getter methods for interfaces
func (c *Container) GetLogger() *shared.CustomLogger {
	return c.Logger
}

func (c *Container) GetConfig() *config.Config {
	return c.Config
}

func (c *Container) GetStateManager() *statemanager.StateManager {
	return c.StateManager
}

func (c *Container) GetCredentials() []auth.Credential {
	return c.Credentials
}

func (c *Container) GetHTTPClient() *http.Client {
	return c.HTTPClient
}

func (c *Container) GetDB() *sql.DB {
	return c.DB
}

func (c *Container) GetDBQueries() *dbgen.Queries {
	return c.DBQueries
}

// Setter methods
func (c *Container) SetLogger(logger *shared.CustomLogger) *Container {
	c.Logger = logger
	return c
}

func (c *Container) SetConfig(config *config.Config) *Container {
	c.Config = config
	return c
}

func (c *Container) SetStateManager(stateManager *statemanager.StateManager) *Container {
	c.StateManager = stateManager
	return c
}

func (c *Container) SetCredentials(credentials []auth.Credential) *Container {
	c.Credentials = credentials
	return c
}

func (c *Container) SetDB(db *sql.DB) *Container {
	c.DB = db
	c.DBQueries = dbgen.New(db)
	return c
}

func (c *Container) SetHTTPClient(client *http.Client) *Container {
	c.HTTPClient = client
	return c
}

// Validate ensures all required dependencies are set
func (c *Container) Validate() error {
	if c.Logger == nil {
		return fmt.Errorf("logger is required")
	}
	if c.Config == nil {
		return fmt.Errorf("config is required")
	}
	if c.StateManager == nil {
		return fmt.Errorf("state manager is required")
	}
	if len(c.Credentials) == 0 {
		return fmt.Errorf("credentials are required")
	}
	if c.DB == nil {
		return fmt.Errorf("database is required")
	}
	if c.HTTPClient == nil {
		return fmt.Errorf("HTTP client is required")
	}
	return nil
}
