package bitbucketcloud

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/integrations/git/gitdtos"
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
		devDRepos := []gitdtos.BLRepo{}
		for _, repo := range repos {
			devDRepos = append(devDRepos, gitdtos.BLRepo{
				Slug:     repo.Slug,
				Name:     repo.Name,
				ID:       repo.ID,
				IsPublic: !repo.IsPrivate,
				Link:     repo.Links.HTML.Href,
				Commits:  []gitdtos.BLCommit{},
				Prs:      []gitdtos.BLPullRequest{},
			})
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
				WorkspaceSlug:      workspace.Slug,
				SuccessfulSyncTime: sql.NullTime{Valid: false},
				Success:            false,
				ErrorContext:       sql.NullString{Valid: false},
			}); err != nil {
				bcSvc.logger.Error("Error creating repo sync audit", "error", err)
				return err
			}
		}

		if err := bcSvc.dataRelayer.SendCollectedData(devDRepos, url.Values{}); err != nil {
			bcSvc.logger.Error("Error sending pull data to data relayer", "error", err)
			return err
		}
	}

	bcSvc.logger.Info("Repositories pulled successfully.")
	return nil
}

func (bcSvc *BitbucketCloudSvc) GitActivityPull() error {
	bcSvc.logger.Info("Pulling Git activity from Bitbucket Cloud...")
	savedRepos, err := bcSvc.getAllActiveRepoSyncAudits()
	if err != nil {
		bcSvc.logger.Error("Error getting all active repo sync audits", "error", err)
		return err
	}
	bcSvc.logger.Info("Found active repo sync audits", "count", len(savedRepos))
	for _, repoSyncAudit := range savedRepos {
		bcSvc.logger.Info("Repo sync audit", "repoName", repoSyncAudit.RepoName)

		currentSyncTime := time.Now()
		if err := bcSvc.syncGitActivityForRepo(repoSyncAudit); err != nil {
			bcSvc.logger.Error("Error syncing Git activity for repo", "error", err)

			repoSyncAudit.Success = false
			repoSyncAudit.ErrorContext = sql.NullString{String: err.Error(), Valid: true}
			repoSyncAudit.UpdatedAt = currentSyncTime
			if _, err := bcSvc.dbQuerier.UpdateRepoSyncAudit(context.Background(), dbgen.UpdateRepoSyncAuditParams{
				ID:                 repoSyncAudit.ID,
				RepoName:           repoSyncAudit.RepoName,
				WorkspaceSlug:      repoSyncAudit.WorkspaceSlug,
				SuccessfulSyncTime: sql.NullTime{Time: currentSyncTime, Valid: true},
				Success:            false,
				ErrorContext:       sql.NullString{String: err.Error(), Valid: true},
			}); err != nil {
				bcSvc.logger.Error("Error updating repo sync audit", "error", err)
				return err
			}
		} else {
			// Send success event to data relayer
			repoSyncAudit.Success = true
			repoSyncAudit.ErrorContext = sql.NullString{Valid: false}
			repoSyncAudit.UpdatedAt = currentSyncTime
			if _, err := bcSvc.dbQuerier.UpdateRepoSyncAudit(context.Background(), dbgen.UpdateRepoSyncAuditParams{
				ID:                 repoSyncAudit.ID,
				RepoName:           repoSyncAudit.RepoName,
				WorkspaceSlug:      repoSyncAudit.WorkspaceSlug,
				SuccessfulSyncTime: sql.NullTime{Time: currentSyncTime, Valid: true},
				Success:            true,
				ErrorContext:       sql.NullString{Valid: false},
			}); err != nil {
				bcSvc.logger.Error("Error updating repo sync audit", "error", err)
				return err
			}
		}
	}

	bcSvc.logger.Info("Git activity pulled successfully.")
	return nil
}

func (bcSvc *BitbucketCloudSvc) getAllActiveRepoSyncAudits() ([]dbgen.RepositorySyncAudit, error) {
	var repoSyncAudits []dbgen.RepositorySyncAudit
	limit := 100
	for {
		repoSyncAuditsPerPage, err := bcSvc.dbQuerier.ListActiveRepoSyncAuditOrderBySuccessfulSyncTimeCreatedAt(context.Background(), dbgen.ListActiveRepoSyncAuditOrderBySuccessfulSyncTimeCreatedAtParams{
			Offset: int64(len(repoSyncAudits)),
			Limit:  100,
		})
		if err != nil {
			return nil, fmt.Errorf("error getting paginated repo sync audits: %w", err)
		}
		repoSyncAudits = append(repoSyncAudits, repoSyncAuditsPerPage...)
		if len(repoSyncAuditsPerPage) < limit {
			break
		}
	}
	return repoSyncAudits, nil
}

func (bcSvc *BitbucketCloudSvc) syncGitActivityForRepo(repoSyncAudit dbgen.RepositorySyncAudit) *gitdtos.BLRepoError {
	repoError := &gitdtos.BLRepoError{}
	{
		var prError map[string]any = map[string]any{}
		fetchedPRs, err := bcSvc.apiClient.GetPullRequestsByRepository(repoSyncAudit.WorkspaceSlug, repoSyncAudit.ID, bcSvc.dataRelayer.SendPullError)
		if err != nil {
			repoError.PrFetchError = err.Error()
		}

		devDPRs := []dtos.BLPullRequest{}
		prErrors := []map[string]any{}
		for _, bBktCloudPr := range fetchedPRs {
			var perPrError map[string]any = map[string]any{}

			fetchedPrCommits, err := bcSvc.apiClient.GetPullRequestCommits(repoSyncAudit.WorkspaceSlug, repoSyncAudit.ID, bBktCloudPr.ID, bcSvc.dataRelayer.SendPullError)
			if err != nil {
				perPrError["commitFetchError"] = err
			}

			devDCommits := []dtos.BLCommit{}
			for _, commit := range fetchedPrCommits {
				devDCommits = append(devDCommits, dtos.BLCommit{
					ID:                 commit.Hash,
					Message:            commit.Message,
					Committer:          convertBBktCloudActorToDevDActor(commit.Author),
					CommitterTimestamp: commit.Date,
					ChangedFiles:       []dtos.BLChangedFile{},
				})
			}

			isOpen := bBktCloudPr.State == string(BBktCloudPullRequestStateOpen)
			reviewers := make([]dtos.BLActor, len(bBktCloudPr.Reviewers))
			for i, reviewer := range bBktCloudPr.Reviewers {
				reviewers[i] = convertBBktCloudActorToDevDActor(reviewer)
			}
			devDPR := dtos.BLPullRequest{
				ID:           bBktCloudPr.ID,
				Title:        bBktCloudPr.Title,
				Description:  bBktCloudPr.Description,
				State:        bBktCloudPr.State,
				Open:         isOpen,
				Closed:       !isOpen,
				CreatedDate:  bBktCloudPr.CreatedOn,
				UpdatedDate:  bBktCloudPr.UpdatedOn,
				SourceBranch: bBktCloudPr.Source.Branch.Name,
				TargetBranch: bBktCloudPr.Destination.Branch.Name,
				Author:       convertBBktCloudActorToDevDActor(bBktCloudPr.Author),
				Reviewers:    reviewers,
				CommentCount: bBktCloudPr.CommentCount,
				Link:         bBktCloudPr.Links.HTML.Href,
				PrCommits:    devDCommits,
			}
			devDPRs = append(devDPRs, devDPR)

			if len(perPrError) > 0 {
				prErrors = append(prErrors, perPrError)
			}
		}
		if len(prErrors) > 0 {
			prError[""] = prErrors
		}
		if len(prError) > 0 {
			repoError["prErrors"] = prError
		}
	}

	bcSvc.logger.Info("Found pull requests", "count", len(fetchedPRs))
	for _, pullRequest := range fetchedPRs {
		bcSvc.logger.Info("Pull request", "title", pullRequest.Title)
	}
	return nil
}

func convertBBktCloudActorToDevDActor(bBktCloudActor BBKtCloudActor) dtos.BLActor {
	return dtos.BLActor{
		ID:           bBktCloudActor.Raw,
		Name:         bBktCloudActor.Raw,
		DisplayName:  bBktCloudActor.Raw,
		EmailAddress: bBktCloudActor.Raw,
	}
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
