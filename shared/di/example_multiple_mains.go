package di

import (
	"database/sql"
	"net/http"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

// Example: Different main files can use different containers based on their needs

// Example 1: datapuller/main.go - needs full container
func ExampleDataPullerMain() {
	// This main needs everything
	container := NewContainer().
		SetLogger(&shared.CustomLogger{}).
		SetConfig(&config.Config{}).
		SetStateManager(&statemanager.StateManager{}).
		SetCredentials([]auth.Credential{{}}).
		SetDB(&sql.DB{}).
		SetHTTPClient(&http.Client{})

	// Use full container
	_ = NewSomeService(container)
}

// Example 2: authsync/main.go - needs only basic dependencies
func ExampleAuthSyncMain() {
	// This main only needs logger, config, and credentials
	serviceContainer := NewServiceContainer().
		SetLogger(&shared.CustomLogger{}).
		SetConfig(&config.Config{}).
		SetCredentials([]auth.Credential{{}})

	// No database, no HTTP client needed
	_ = NewAuthSyncService(serviceContainer)
}

// Example 3: api/main.go - needs API-related dependencies
func ExampleAPIMain() {
	// This main needs API-related dependencies
	apiContainer := NewAPIContainer().
		SetLogger(&shared.CustomLogger{}).
		SetStateManager(&statemanager.StateManager{}).
		SetCredentials([]auth.Credential{{}}).
		SetHTTPClient(&http.Client{})

	// No config, no database needed
	_ = NewAPIService(apiContainer)
}

// Example 4: worker/main.go - needs database and basic dependencies
func ExampleWorkerMain() {
	// This main needs database and basic dependencies
	serviceContainer := NewServiceContainer().
		SetLogger(&shared.CustomLogger{}).
		SetConfig(&config.Config{}).
		SetStateManager(&statemanager.StateManager{}).
		SetCredentials([]auth.Credential{{}})

	databaseContainer := NewDatabaseContainer().
		SetDB(&sql.DB{})

	// Pass only what's needed
	_ = NewWorkerService(serviceContainer, databaseContainer)
}

// Mock service constructors for the examples
func NewSomeService(container *Container) interface{} {
	return nil
}

func NewAuthSyncService(serviceProvider ServiceProvider) interface{} {
	return nil
}

func NewAPIService(apiProvider APIProvider) interface{} {
	return nil
}

func NewWorkerService(serviceProvider ServiceProvider, databaseProvider DatabaseProvider) interface{} {
	return nil
}
