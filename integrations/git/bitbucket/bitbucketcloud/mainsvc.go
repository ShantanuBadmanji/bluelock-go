package bitbucketcloud

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
	dbgen "github.com/bluelock-go/shared/database/generated"
	"github.com/bluelock-go/shared/di"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

type BitbucketCloudSvc struct {
	serviceProvider  di.ServiceProvider
	databaseProvider di.DatabaseProvider
	apiClient        *Client
}

func NewBitbucketCloudSvc(serviceProvider di.ServiceProvider, apiProvider di.APIProvider, databaseProvider di.DatabaseProvider) *BitbucketCloudSvc {
	return &BitbucketCloudSvc{
		serviceProvider:  serviceProvider,
		databaseProvider: databaseProvider,
		apiClient:        NewClient(apiProvider),
	}
}

// Implement ServiceProvider interface methods (required for IntegrationService)
func (bcSvc *BitbucketCloudSvc) GetLogger() *shared.CustomLogger {
	return bcSvc.serviceProvider.GetLogger()
}

func (bcSvc *BitbucketCloudSvc) GetConfig() *config.Config {
	return bcSvc.serviceProvider.GetConfig()
}

func (bcSvc *BitbucketCloudSvc) GetStateManager() *statemanager.StateManager {
	return bcSvc.serviceProvider.GetStateManager()
}

func (bcSvc *BitbucketCloudSvc) GetCredentials() []auth.Credential {
	return bcSvc.serviceProvider.GetCredentials()
}

func (bcSvc *BitbucketCloudSvc) ValidateEnvVariables() error {
	bcSvc.serviceProvider.GetLogger().Info("Validating environment variables for Bitbucket Cloud...")

	BitbucketCloudConfig := bcSvc.serviceProvider.GetConfig().Integrations.BitbucketCloud
	if BitbucketCloudConfig.Workspace == "" {
		return fmt.Errorf("bitbucket Cloud workspace is not set in the configuration")
	}

	return nil
}

func (bcSvc *BitbucketCloudSvc) RunJob() error {
	bcSvc.serviceProvider.GetLogger().Info("Bitbucket Cloud job started...")

	if err := bcSvc.RepoPull(); err != nil {
		bcSvc.serviceProvider.GetLogger().Error("Error pulling repositories from Bitbucket Cloud", "error", err)
		return err
	}

	if err := bcSvc.GitActivityPull(); err != nil {
		bcSvc.serviceProvider.GetLogger().Error("Error pulling Git activity from Bitbucket Cloud", "error", err)
		return err
	}

	time.Sleep(time.Second * 5)
	bcSvc.serviceProvider.GetLogger().Info("Bitbucket Cloud job completed.")
	return nil
}

// RepoPull implements GitIntegrationService.RepoPull()
func (bcSvc *BitbucketCloudSvc) RepoPull() error {
	bcSvc.serviceProvider.GetLogger().Info("Pulling repositories from Bitbucket Cloud...")
	workspaces, err := bcSvc.apiClient.GetWorkspaces(func(s string) {})
	if err != nil {
		bcSvc.serviceProvider.GetLogger().Error("Error pulling workspaces from Bitbucket Cloud", "error", err)
		return err
	}
	if len(workspaces) == 0 {
		return fmt.Errorf("no workspaces found in Bitbucket Cloud")
	}
	bcSvc.serviceProvider.GetLogger().Info("Found workspaces", "count", len(workspaces))
	for _, workspace := range workspaces {
		repos, err := bcSvc.apiClient.GetRepositoriesByWorkspace(workspace.Slug, func(s string) {})
		if err != nil {
			bcSvc.serviceProvider.GetLogger().Error("Error pulling repositories from Bitbucket Cloud", "error", err)
			return err
		}
		if len(repos) == 0 {
			bcSvc.serviceProvider.GetLogger().Warn("No repositories found in workspace", "workspace", workspace.Slug)
			continue
		}
		bcSvc.serviceProvider.GetLogger().Info("Found repositories", "count", len(repos))
		for _, repo := range repos {
			bcSvc.serviceProvider.GetLogger().Info("Repository", "name", repo.Name)
			if repoSyncAudit, err := bcSvc.databaseProvider.GetDBQueries().GetRepoSyncAuditByID(context.Background(), repo.Slug); err != nil {
				bcSvc.serviceProvider.GetLogger().Error("Error getting repo sync audit", "error", err)
				return err
			} else if repoSyncAudit.ID != "" {
				bcSvc.serviceProvider.GetLogger().Info("Repo sync audit already exists", "id", repoSyncAudit.ID)
				continue
			}

			if _, err := bcSvc.databaseProvider.GetDBQueries().CreateRepoSyncAudit(context.Background(), dbgen.CreateRepoSyncAuditParams{
				ID:                 repo.Slug,
				RepoName:           repo.Name,
				SuccessfulSyncTime: sql.NullTime{Valid: false},
				Success:            false,
				ErrorContext:       sql.NullString{Valid: false},
			}); err != nil {
				bcSvc.serviceProvider.GetLogger().Error("Error creating repo sync audit", "error", err)
				return err
			}

		}
	}

	// Simulate pulling repositories
	// In a real implementation, this would involve making API calls to Bitbucket Cloud
	// to pull the repositories and store them in the state manager.
	bcSvc.serviceProvider.GetLogger().Info("Repositories pulled successfully.")
	return nil
}

// GitActivityPull implements GitIntegrationService.GitActivityPull()
func (bcSvc *BitbucketCloudSvc) GitActivityPull() error {
	bcSvc.serviceProvider.GetLogger().Info("Pulling Git activity from Bitbucket Cloud...")
	// Simulate pulling Git activity
	// In a real implementation, this would involve making API calls to Bitbucket Cloud
	// to pull the Git activity and store them in the state manager.
	bcSvc.serviceProvider.GetLogger().Info("Git activity pulled successfully.")
	return nil
}
