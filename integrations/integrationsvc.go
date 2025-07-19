package integrations

import (
	"fmt"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/integrations/git/bitbucket/bitbucketcloud"
	"github.com/bluelock-go/shared/di"
)

type IntegrationService interface {
	di.ServiceProvider
	// ValidateEnvVariables validates the environment variables for the integration service.
	ValidateEnvVariables() error
	// RunJob runs the data pull for the integration service.
	RunJob() error
}

// GetActiveIntegrationService returns an active integration service implementation based on the configuration in the provided container.
// If the configured service type is unsupported, it returns an error.
func GetActiveIntegrationService(container *di.Container) (IntegrationService, error) {
	var service IntegrationService
	var err error

	switch container.Config.ActiveService {
	case config.BitbucketCloudKey:
		container.Logger.Info("Initializing Bitbucket Cloud as the active integration service")

		// Create specialized containers for different needs
		serviceContainer := container.ToServiceContainer()
		apiContainer := container.ToAPIContainer()
		databaseContainer := container.ToDatabaseContainer()

		service = bitbucketcloud.NewBitbucketCloudSvc(serviceContainer, apiContainer, databaseContainer)
	default:
		container.Logger.Error("Unsupported service type", "serviceType", container.Config.ActiveService)
		return nil, fmt.Errorf("unsupported service type: %s", container.Config.ActiveService)
	}

	return service, err
}
