package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/integrations/bitbucket/bitbucketcloud"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/jobscheduler"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

const (
	cronExpr = "0 * * * *" // Every hour
)

func main() {
	appLoggerFilePath := filepath.Join(shared.RootDir, "logs", "datapuller.log")

	customLogger, logFile, err := shared.NewCustomLogger(appLoggerFilePath, shared.TextLogHandler)
	if err != nil {
		log.Fatalf("failed to create custom logger: %v", err)
	}
	customLogger.Info("Custom logger initialized", "absoluteFilePath", appLoggerFilePath)

	defer logFile.Close()

	// Start the job scheduler
	stateJsonFilePath := filepath.Join(shared.RootDir, "states", "datapuller.json")
	stateManager, err := statemanager.NewStateManager(stateJsonFilePath)
	if err != nil {
		customLogger.Logger.Error("Failed to initialize state manager", "error", err)
		os.Exit(1)
	}

	// initialte services
	bitbucketcloudSvc := bitbucketcloud.NewBitbucketCloudSvc(customLogger, stateManager)

	// Initialize the job scheduler
	scheduler, err := jobscheduler.NewJobScheduler(customLogger, stateManager, "datapuller", bitbucketcloudSvc.RunJob, &config.Config{CronExpr: cronExpr})
	if err != nil {
		customLogger.Error("Failed to initialize job scheduler", "error", err)
		os.Exit(1)
	}
	customLogger.Info("Job scheduler initialized")

	scheduler.Run()
	customLogger.Info("Job scheduler stopped")
	customLogger.Info("Exiting application...")
}
