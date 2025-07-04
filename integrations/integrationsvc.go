package integrations

import (
	"fmt"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/integrations/git/bitbucket/bitbucketcloud"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

type IntegrationService interface {
	// GetLogger returns the logger of the integration service.
	GetLogger() *shared.CustomLogger
	// GetConfig returns the configuration of the integration service.
	GetConfig() *config.Config
	// GetCredentials returns the credentials of the integration service.
	GetCredentials() []auth.Credential
	// GetStateManager returns the state manager of the integration service.
	GetStateManager() *statemanager.StateManager
	// ValidateEnvVariables validates the environment variables for the integration service.
	ValidateEnvVariables() error
	// DataPull runs the data pull for the integration service.
	RunJob() error
}

func GetActiveIntegrationService(cfg *config.Config, logger *shared.CustomLogger, stateManager *statemanager.StateManager, credentials []auth.Credential) (IntegrationService, error) {
	switch cfg.ActiveService {
	case config.BitbucketCloudKey:
		logger.Info("Initializing Bitbucket Cloud as the active integration service")
		return bitbucketcloud.NewBitbucketCloudSvc(logger, stateManager, credentials, cfg), nil
	default:
		logger.Error("Unsupported service type", "serviceType", cfg.ActiveService)
		return nil, fmt.Errorf("unsupported service type: %s", cfg.ActiveService)
	}
}
