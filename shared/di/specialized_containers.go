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

// APIContainer contains only dependencies needed for API services
type APIContainer struct {
	Logger       *shared.CustomLogger
	StateManager *statemanager.StateManager
	Credentials  []auth.Credential
	HTTPClient   *http.Client
}

// Ensure APIContainer implements APIProvider
var _ APIProvider = (*APIContainer)(nil)

func NewAPIContainer() *APIContainer {
	return &APIContainer{
		HTTPClient: http.DefaultClient,
	}
}

func (ac *APIContainer) GetLogger() *shared.CustomLogger {
	return ac.Logger
}

func (ac *APIContainer) GetStateManager() *statemanager.StateManager {
	return ac.StateManager
}

func (ac *APIContainer) GetCredentials() []auth.Credential {
	return ac.Credentials
}

func (ac *APIContainer) GetHTTPClient() *http.Client {
	return ac.HTTPClient
}

func (ac *APIContainer) SetLogger(logger *shared.CustomLogger) *APIContainer {
	ac.Logger = logger
	return ac
}

func (ac *APIContainer) SetStateManager(stateManager *statemanager.StateManager) *APIContainer {
	ac.StateManager = stateManager
	return ac
}

func (ac *APIContainer) SetCredentials(credentials []auth.Credential) *APIContainer {
	ac.Credentials = credentials
	return ac
}

func (ac *APIContainer) SetHTTPClient(client *http.Client) *APIContainer {
	ac.HTTPClient = client
	return ac
}

// Validate ensures all required dependencies for API services are set
func (ac *APIContainer) Validate() error {
	if ac.Logger == nil {
		return fmt.Errorf("logger is required for API services")
	}
	if ac.StateManager == nil {
		return fmt.Errorf("state manager is required for API services")
	}
	if len(ac.Credentials) == 0 {
		return fmt.Errorf("credentials are required for API services")
	}
	if ac.HTTPClient == nil {
		return fmt.Errorf("HTTP client is required for API services")
	}
	return nil
}

// ServiceContainer contains only dependencies needed for main services
type ServiceContainer struct {
	Logger       *shared.CustomLogger
	Config       *config.Config
	StateManager *statemanager.StateManager
	Credentials  []auth.Credential
}

// Ensure ServiceContainer implements ServiceProvider
var _ ServiceProvider = (*ServiceContainer)(nil)

func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{}
}

func (sc *ServiceContainer) GetLogger() *shared.CustomLogger {
	return sc.Logger
}

func (sc *ServiceContainer) GetConfig() *config.Config {
	return sc.Config
}

func (sc *ServiceContainer) GetStateManager() *statemanager.StateManager {
	return sc.StateManager
}

func (sc *ServiceContainer) GetCredentials() []auth.Credential {
	return sc.Credentials
}

func (sc *ServiceContainer) SetLogger(logger *shared.CustomLogger) *ServiceContainer {
	sc.Logger = logger
	return sc
}

func (sc *ServiceContainer) SetConfig(config *config.Config) *ServiceContainer {
	sc.Config = config
	return sc
}

func (sc *ServiceContainer) SetStateManager(stateManager *statemanager.StateManager) *ServiceContainer {
	sc.StateManager = stateManager
	return sc
}

func (sc *ServiceContainer) SetCredentials(credentials []auth.Credential) *ServiceContainer {
	sc.Credentials = credentials
	return sc
}

// Validate ensures all required dependencies for main services are set
func (sc *ServiceContainer) Validate() error {
	if sc.Logger == nil {
		return fmt.Errorf("logger is required for main services")
	}
	if sc.Config == nil {
		return fmt.Errorf("config is required for main services")
	}
	if sc.StateManager == nil {
		return fmt.Errorf("state manager is required for main services")
	}
	if len(sc.Credentials) == 0 {
		return fmt.Errorf("credentials are required for main services")
	}
	return nil
}

// DatabaseContainer contains only database-related dependencies
type DatabaseContainer struct {
	DB        *sql.DB
	DBQueries *dbgen.Queries
}

// Ensure DatabaseContainer implements DatabaseProvider
var _ DatabaseProvider = (*DatabaseContainer)(nil)

func NewDatabaseContainer() *DatabaseContainer {
	return &DatabaseContainer{}
}

func (dc *DatabaseContainer) GetDB() *sql.DB {
	return dc.DB
}

func (dc *DatabaseContainer) GetDBQueries() *dbgen.Queries {
	return dc.DBQueries
}

func (dc *DatabaseContainer) SetDB(db *sql.DB) *DatabaseContainer {
	dc.DB = db
	dc.DBQueries = dbgen.New(db)
	return dc
}

// Validate ensures all required dependencies for database operations are set
func (dc *DatabaseContainer) Validate() error {
	if dc.DB == nil {
		return fmt.Errorf("database is required for database operations")
	}
	if dc.DBQueries == nil {
		return fmt.Errorf("database queries are required for database operations")
	}
	return nil
}

// Factory functions to create specialized containers from full container
func (c *Container) ToAPIContainer() *APIContainer {
	return NewAPIContainer().
		SetLogger(c.Logger).
		SetStateManager(c.StateManager).
		SetCredentials(c.Credentials).
		SetHTTPClient(c.HTTPClient)
}

func (c *Container) ToServiceContainer() *ServiceContainer {
	return NewServiceContainer().
		SetLogger(c.Logger).
		SetConfig(c.Config).
		SetStateManager(c.StateManager).
		SetCredentials(c.Credentials)
}

func (c *Container) ToDatabaseContainer() *DatabaseContainer {
	return NewDatabaseContainer().
		SetDB(c.DB)
}
