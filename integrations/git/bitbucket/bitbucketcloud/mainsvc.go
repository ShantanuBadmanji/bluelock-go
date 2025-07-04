package bitbucketcloud

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

type BitbucketCloudSvc struct {
	logger       *shared.CustomLogger
	stateManager *statemanager.StateManager
	credentials  []auth.Credential
	config       *config.Config
	apiClient    *Client
}

func NewBitbucketCloudSvc(logger *shared.CustomLogger, stateManager *statemanager.StateManager, credentials []auth.Credential, config *config.Config) *BitbucketCloudSvc {
	return &BitbucketCloudSvc{logger, stateManager, credentials, config,
		NewClient(http.DefaultClient, stateManager, logger, credentials),
	}
}

func (bcSvc *BitbucketCloudSvc) GetLogger() *shared.CustomLogger {
	return bcSvc.logger
}
func (bcSvc *BitbucketCloudSvc) GetConfig() *config.Config {
	return bcSvc.config
}
func (bcSvc *BitbucketCloudSvc) GetStateManager() *statemanager.StateManager {
	return bcSvc.stateManager
}
func (bcSvc *BitbucketCloudSvc) GetCredentials() []auth.Credential {
	return bcSvc.credentials
}

func (bcSvc *BitbucketCloudSvc) ValidateEnvVariables() error {
	bcSvc.logger.Info("Validating environment variables for Bitbucket Cloud...")

	BitbucketCloudConfig := bcSvc.config.Integrations.BitbucketCloud
	if BitbucketCloudConfig.Workspace == "" {
		return fmt.Errorf("bitbucket Cloud workspace is not set in the configuration")
	}

	return nil
}

func (bcSvc *BitbucketCloudSvc) RunJob() error {
	bcSvc.logger.Info("Bitbucket Cloud job started...")

	if err := bcSvc.RepoPull(); err != nil {
		bcSvc.logger.Error("Error pulling repositories from Bitbucket Cloud", "error", err)
		return err
	}

	if err := bcSvc.GitActivityPull(); err != nil {
		bcSvc.logger.Error("Error pulling Git activity from Bitbucket Cloud", "error", err)
		return err
	}

	time.Sleep(time.Second * 5)
	bcSvc.logger.Info("Bitbucket Cloud job completed.")
	return nil
}

func (bcSvc *BitbucketCloudSvc) RepoPull() error {
	bcSvc.logger.Info("Pulling repositories from Bitbucket Cloud...")
	workspaces, err := bcSvc.apiClient.GetWorkspaces(func(s string) {})
	if err != nil {
		bcSvc.logger.Error("Error pulling workspaces from Bitbucket Cloud", "error", err)
		return err
	}
	if len(workspaces) == 0 {
		return fmt.Errorf("no workspaces found in Bitbucket Cloud")
	}
	bcSvc.logger.Info("Found workspaces", "count", len(workspaces))
	for _, workspace := range workspaces {

		bcSvc.logger.Info("Processing workspace", "slug", workspace.Slug, "name", workspace.Name)
	}

	// Simulate pulling repositories
	// In a real implementation, this would involve making API calls to Bitbucket Cloud
	// to pull the repositories and store them in the state manager.
	bcSvc.logger.Info("Repositories pulled successfully.")
	return nil
}

func (bcSvc *BitbucketCloudSvc) GitActivityPull() error {
	bcSvc.logger.Info("Pulling Git activity from Bitbucket Cloud...")
	// Simulate pulling Git activity
	// In a real implementation, this would involve making API calls to Bitbucket Cloud
	// to pull the Git activity and store them in the state manager.
	bcSvc.logger.Info("Git activity pulled successfully.")
	return nil
}
