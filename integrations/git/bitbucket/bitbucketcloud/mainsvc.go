package bitbucketcloud

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/integrations/relay"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
	"github.com/bluelock-go/shared/auth/credservice"
	"github.com/bluelock-go/shared/database/dbsetup"
	dbgen "github.com/bluelock-go/shared/database/generated"
	"github.com/bluelock-go/shared/di"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

type BitbucketCloudSvc struct {
	logger       *shared.CustomLogger
	stateManager *statemanager.StateManager
	credentials  []auth.Credential
	config       *config.Config
	apiClient    *Client
	dbQuerier    dbgen.Querier
	dataRelayer  relay.DataRelayer
}

func NewBitbucketCloudSvc(logger *shared.CustomLogger, stateManager *statemanager.StateManager, credentials []auth.Credential, config *config.Config, dbQuerier dbgen.Querier, client *Client, dataRelayer relay.DataRelayer) *BitbucketCloudSvc {
	return &BitbucketCloudSvc{logger, stateManager, credentials, config,
		client,
		dbQuerier,
		dataRelayer,
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
func (bcSvc *BitbucketCloudSvc) GetQuerier() dbgen.Querier {
	return bcSvc.dbQuerier
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
	workspaces, err := bcSvc.apiClient.GetWorkspaces(bcSvc.dataRelayer.SendPullError)
	if err != nil {
		bcSvc.logger.Error("Error pulling workspaces from Bitbucket Cloud", "error", err)
		return err
	}
	if len(workspaces) == 0 {
		return fmt.Errorf("no workspaces found in Bitbucket Cloud")
	}
	bcSvc.logger.Info("Found workspaces", "count", len(workspaces))
	for _, workspace := range workspaces {
		repos, err := bcSvc.apiClient.GetRepositoriesByWorkspace(workspace.Slug, bcSvc.dataRelayer.SendPullError)
		if err != nil {
			bcSvc.logger.Error("Error pulling repositories from Bitbucket Cloud", "error", err)
			return err
		}
		if len(repos) == 0 {
			bcSvc.logger.Warn("No repositories found in workspace", "workspace", workspace.Slug)
			continue
		}
		bcSvc.logger.Info("Found repositories", "count", len(repos))
		for _, repo := range repos {
			bcSvc.logger.Info("Repository", "name", repo.Name)
			if existingRepoSyncAudit, err := bcSvc.dbQuerier.GetRepoSyncAuditByID(context.Background(), repo.Slug); err == nil || errors.Is(err, sql.ErrNoRows) {
				bcSvc.logger.Debug("Repository found in database", "name", existingRepoSyncAudit.RepoName)
				continue
			} else if errors.Is(err, sql.ErrNoRows) {
				// If the error is a no rows error, create a new repo sync audit
				bcSvc.logger.Info("Repository not found in database. Creating new repo sync audit", "name", repo.Name)
			} else {
				bcSvc.logger.Error("Error getting repo sync audit", "error", err)
				return err
			}

			if _, err := bcSvc.dbQuerier.CreateRepoSyncAudit(context.Background(), dbgen.CreateRepoSyncAuditParams{
				ID:                 repo.Slug,
				RepoName:           repo.Name,
				SuccessfulSyncTime: sql.NullTime{Valid: false},
				Success:            false,
				ErrorContext:       sql.NullString{Valid: false},
			}); err != nil {
				bcSvc.logger.Error("Error creating repo sync audit", "error", err)
				return err
			}
		}
	}

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

var bitbucketCloudSvc = di.NewThreadSafeSingleton(func() *BitbucketCloudSvc {
	customLogger := shared.AcquireCustomLogger()
	cfg := config.AcquireConfig()
	statemanager := statemanager.AcquireStateManager()
	credentials := credservice.AcquireCredentials()
	dbQuerier := dbsetup.AcquireQuerier()
	client := AcquireClient()
	bluelockDataRelayer := relay.AcquireBluelockRelayService()
	return NewBitbucketCloudSvc(customLogger, statemanager, credentials, cfg, dbQuerier, client, bluelockDataRelayer)
})

func AcquireBitbucketCloudSvc() *BitbucketCloudSvc {
	return bitbucketCloudSvc.Acquire()
}
