package bitbucketcloud

import (
	"fmt"
	"time"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

type BitbucketCloudSvc struct {
	logger       *shared.CustomLogger
	stateManager *statemanager.StateManager
	config       *config.Config
}

func NewBitbucketCloudSvc(logger *shared.CustomLogger, stateManager *statemanager.StateManager, config *config.Config) *BitbucketCloudSvc {
	return &BitbucketCloudSvc{logger, stateManager, config}
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
	time.Sleep(time.Second * 5)
	bcSvc.logger.Info("Bitbucket Cloud job completed.")
	return nil
}
