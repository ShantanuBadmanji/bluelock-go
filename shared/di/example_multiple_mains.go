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

// ExampleDataPullerMain demonstrates initializing a full dependency injection container with all available dependencies and using it to construct a service.
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

// ExampleAuthSyncMain demonstrates initializing a service container with only the logger, config, and credentials dependencies, suitable for applications that do not require a database or HTTP client.
func ExampleAuthSyncMain() {
	// This main only needs logger, config, and credentials
	serviceContainer := NewServiceContainer().
		SetLogger(&shared.CustomLogger{}).
		SetConfig(&config.Config{}).
		SetCredentials([]auth.Credential{{}})

	// No database, no HTTP client needed
	_ = NewAuthSyncService(serviceContainer)
}

// ExampleAPIMain demonstrates initializing an API service with only the dependencies required for API operations, such as logger, state manager, credentials, and HTTP client, using a tailored dependency injection container.
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

// ExampleWorkerMain demonstrates initializing a worker service with separate service and database containers, each configured with only the required dependencies.
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

// NewSomeService returns a placeholder service instance using the provided Container.
// This function serves as a mock constructor for demonstration purposes.
func NewSomeService(container *Container) interface{} {
	return nil
}

// NewAuthSyncService creates a new instance of the auth sync service using the provided service dependencies.
func NewAuthSyncService(serviceProvider ServiceProvider) interface{} {
	return nil
}

// NewAPIService initializes a new API service using the provided APIProvider.
// Returns a placeholder value; replace with actual service initialization as needed.
func NewAPIService(apiProvider APIProvider) interface{} {
	return nil
}

// NewWorkerService initializes a worker service using the provided service and database dependencies.
// Returns a placeholder value; intended as a stub for actual worker service construction.
func NewWorkerService(serviceProvider ServiceProvider, databaseProvider DatabaseProvider) interface{} {
	return nil
}
