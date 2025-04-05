package integrations

import (
	"github.com/bluelock-go/config"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

type IntegrationService interface {
	// GetLogger returns the logger of the integration service.
	GetLogger() *shared.CustomLogger
	// GetConfig returns the configuration of the integration service.
	GetConfig() *config.Config
	// GetStateManager returns the state manager of the integration service.
	GetStateManager() *statemanager.StateManager
	// ValidateEnvVariables validates the environment variables for the integration service.
	ValidateEnvVariables() error
	// RunJob runs the job for the integration service.
	RunJob() error
}
