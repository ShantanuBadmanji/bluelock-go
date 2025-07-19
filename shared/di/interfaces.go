package di

import (
	"database/sql"
	"net/http"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
	dbgen "github.com/bluelock-go/shared/database/generated"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

// LoggerProvider provides logging capabilities
type LoggerProvider interface {
	GetLogger() *shared.CustomLogger
}

// StateManagerProvider provides state management capabilities
type StateManagerProvider interface {
	GetStateManager() *statemanager.StateManager
}

// CredentialsProvider provides authentication credentials
type CredentialsProvider interface {
	GetCredentials() []auth.Credential
}

// HTTPClientProvider provides HTTP client capabilities
type HTTPClientProvider interface {
	GetHTTPClient() *http.Client
}

// ConfigProvider provides configuration access
type ConfigProvider interface {
	GetConfig() *config.Config
}

// DatabaseProvider provides database access
type DatabaseProvider interface {
	GetDB() *sql.DB
	GetDBQueries() *dbgen.Queries
}

// APIProvider combines dependencies needed for API services
type APIProvider interface {
	LoggerProvider
	StateManagerProvider
	CredentialsProvider
	HTTPClientProvider
}

// ServiceProvider combines dependencies needed for main services
type ServiceProvider interface {
	LoggerProvider
	ConfigProvider
	StateManagerProvider
	CredentialsProvider
}

// FullProvider combines all dependencies (for main functions)
type FullProvider interface {
	LoggerProvider
	ConfigProvider
	StateManagerProvider
	CredentialsProvider
	HTTPClientProvider
	DatabaseProvider
}
