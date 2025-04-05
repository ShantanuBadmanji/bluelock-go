package bitbucketcloud

import (
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

func (bcSvc *BitbucketCloudSvc) ValidateEnvVariables() error {
	bcSvc.logger.Info("Validating environment variables for Bitbucket Cloud...")
	// Add validation logic here
	return nil
}

func (bcSvc *BitbucketCloudSvc) RunJob() {
	bcSvc.logger.Info("Bitbucket Cloud job started...")
	time.Sleep(time.Second * 5)
	bcSvc.logger.Info("Bitbucket Cloud job completed.")
}
