package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/integrations"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth/credservice"
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
		customLogger.Logger.Error("Failed to load authentication tokens", "error", err)
		os.Exit(1)
	} else {
		customLogger.Info("Authentication tokens loaded successfully", "authTokensFilePath", authTokensFilePath)
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
	if err := stateManager.SyncTokenStatusWithLatestAuthCredentials(credStore[credservice.DatapullCredentialsKey]); err != nil {
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

	// initialte services
	customLogger.Info("Initializing Services...")
	customLogger.Info("Initializing Datapull Integration Service...")
	datapullIntegrationSvc, err := integrations.GetActiveIntegrationService(cfg, customLogger, stateManager)
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
