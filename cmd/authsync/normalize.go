// Make sure the all the services using auth_tokens.json are terminated before trying to normalize auth_tokens.json
// as this might lead to corrupted reads, race conditions, or other unexpected behavior.
// This is especially important for services that rely on the auth_tokens.json file for authentication or authorization.
package main

import (
	"log"
	"path/filepath"

	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth/credservice"
)

func main() {
	// Initialize the application logger
	log.Println("Initializing application logger...")
	appLoggerFilePath := filepath.Join(shared.RootDir, "logs", "authsync.log")
	customLogger, logFile, err := shared.NewCustomLogger(appLoggerFilePath, shared.TextLogHandler)
	if err != nil {
		log.Fatalf("failed to create custom logger: %v", err)
	}
	customLogger.Info("Custom logger initialized", "absoluteFilePath", appLoggerFilePath)

	defer logFile.Close()

	// Load authentication tokens
	customLogger.Info("Loading authentication tokens...")
	authTokensFilePath := filepath.Join(shared.RootDir, "secrets", "auth_tokens.json")
	if _, err := credservice.NormalizeAndPersistCredentials(authTokensFilePath); err != nil {
		customLogger.Logger.Error("Failed to normalize and persist credentials", "error", err)
	} else {
		customLogger.Info("Credentials normalized and persisted successfully")
	}
}
