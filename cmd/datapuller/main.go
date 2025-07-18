package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/integrations"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
	"github.com/bluelock-go/shared/auth/credservice"
	dbgen "github.com/bluelock-go/shared/database/generated"
	"github.com/bluelock-go/shared/jobscheduler"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

func main() {
	// Initialize the application logger
	log.Println("Initializing application logger...")
	appLoggerFilePath := filepath.Join(shared.RootDir, "logs", "datapuller.log")
	customLogger, logFile, err := shared.NewCustomLogger(appLoggerFilePath, shared.TextLogHandler)
	if err != nil {
		log.Fatalf("failed to create custom logger: %v", err)
	}
	customLogger.Info("Custom logger initialized", "absoluteFilePath", appLoggerFilePath)

	defer logFile.Close()

	// Load authentication tokens
	customLogger.Info("Loading authentication tokens...")
	authTokensFilePath := filepath.Join(shared.RootDir, "secrets", "auth_tokens.json")
	credStore, _, err := credservice.LoadAuthTokensFromFileAndValidate(authTokensFilePath)
	if err != nil {
		customLogger.Error("Failed to load authentication tokens", "error", err)
		os.Exit(1)
	} else {
		customLogger.Info("Authentication tokens loaded successfully", "authTokensFilePath", authTokensFilePath)
	}

	datapullCredentials, ok := credStore[credservice.DatapullCredentialsKey]
	if !ok {
		customLogger.Error("Datapull credentials not found in the credential store")
		os.Exit(1)
	} else if err := auth.ValidateCredentials(credservice.DatapullCredentialsKey, datapullCredentials); err != nil {
		customLogger.Error("Invalid Datapull credentials", "error", err)
		os.Exit(1)
	} else if len(datapullCredentials) == 0 {
		customLogger.Error("No Datapull credentials found in the credential store")
		os.Exit(1)
	} else {
		customLogger.Info("Datapull credentials found in the credential store", "credentials", datapullCredentials)
	}

	// Initialize the state manager
	customLogger.Info("Initializing state manager...")
	stateJsonFilePath := filepath.Join(shared.RootDir, "states", "datapuller.json")
	stateManager, err := statemanager.NewStateManager(stateJsonFilePath)
	if err != nil {
		customLogger.Logger.Error("Failed to initialize state manager", "error", err)
		os.Exit(1)
	} else {
		customLogger.Info("State manager initialized", "stateJsonFilePath", stateJsonFilePath)
	}

	// Sync token status with the latest authentication credentials
	customLogger.Info("Syncing token status with latest authentication credentials...")
	if err := stateManager.SyncTokenStatusWithLatestAuthCredentials(datapullCredentials); err != nil {
		customLogger.Logger.Error("Failed to sync token status with latest authentication credentials", "error", err)
		os.Exit(1)
	} else {
		customLogger.Info("Token status synced with latest authentication credentials successfully")
	}

	// Load and validate the configuration
	customLogger.Info("Loading configuration...")
	cfg, err := config.LoadMergedConfig()
	if err != nil {
		customLogger.Logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	} else {
		customLogger.Info("Configuration loaded successfully")
	}

	customLogger.Info("Validating defaults and common configuration...")
	if err = cfg.ValidateDefaultsAndCommonConfig(); err != nil {
		customLogger.Logger.Error("Invalid defaults or common configuration", "error", err)
		os.Exit(1)
	} else {
		customLogger.Info("Defaults and common configuration validated successfully")
	}

	//Initialise SQLC DB
	customLogger.Info("Initializing SQLC DB...")
	db, err := sql.Open("sqlite3", filepath.Join(shared.RootDir, "database.db"))
	if err != nil {
		customLogger.Logger.Error("Failed to initialize SQLC DB", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize the database queries
	dbQueries := dbgen.New(db)
	customLogger.Info("SQLC DB initialized successfully", "dbPath", filepath.Join(shared.RootDir, "database.db"))

	// initialte services
	customLogger.Info("Initializing Services...")
	customLogger.Info("Initializing Datapull Integration Service...")
	datapullIntegrationSvc, err := integrations.GetActiveIntegrationService(cfg, customLogger, stateManager, datapullCredentials, dbQueries)
	if err != nil {
		customLogger.Logger.Error("Failed to initialize active integration service", "error", err)
		os.Exit(1)
	}
	customLogger.Info("Datapull Integration Service initialized successfully")

	customLogger.Info("Validating environment variables for integration service...")
	if err := datapullIntegrationSvc.ValidateEnvVariables(); err != nil {
		customLogger.Logger.Error("Failed to validate environment variables for integration service", "error", err)
		os.Exit(1)
	} else {
		customLogger.Info("Environment variables validated successfully for integration service")
	}
	customLogger.Info("Initialized All Services Successfully")

	// Initialize the job scheduler
	scheduler, err := jobscheduler.NewJobScheduler(customLogger, stateManager, "Datapull", datapullIntegrationSvc.RunJob, cfg)
	if err != nil {
		customLogger.Error("Failed to initialize job scheduler", "error", err)
		os.Exit(1)
	} else {
		customLogger.Info("Job scheduler initialized successfully")
	}

	// Start the job scheduler
	customLogger.Info("Starting job scheduler...")
	scheduler.Run()
	customLogger.Info("Job scheduler stopped")
	customLogger.Info("Exiting application...")
}
