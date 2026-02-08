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
	"github.com/bluelock-go/shared/customerrors"
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
		wrappedErr := fmt.Errorf("error pulling repositories from Bitbucket Cloud: %w", err)
		bcSvc.logger.Error(wrappedErr.Error())
		if len(err.CriticalErrors) > 0 {
			return wrappedErr
		}
	}

	if err := bcSvc.GitActivityPull(); err != nil {
		wrappedErr := fmt.Errorf("error pulling Git activity from Bitbucket Cloud: %w", err)
		bcSvc.logger.Error(wrappedErr.Error())
		if len(err.CriticalErrors) > 0 {
			return wrappedErr
		}
	}

	time.Sleep(time.Second * 5)
	bcSvc.logger.Info("Bitbucket Cloud job completed.")
	return nil
}

func (bcSvc *BitbucketCloudSvc) RepoPull() *gitdtos.BLRootErrorPayload {
	rootErrorPayload := &gitdtos.BLRootErrorPayload{}
	bcSvc.logger.Info("Pulling repositories from Bitbucket Cloud...")
	workspaces, err := bcSvc.apiClient.GetWorkspaces(bcSvc.dataRelayer.SendPullError)
	if err != nil {
		wrappedErr := fmt.Errorf("error pulling workspaces from Bitbucket Cloud: %w", err)
		bcSvc.logger.Error(wrappedErr.Error())
		if errors.Is(err, customerrors.ErrCritical) {
			rootErrorPayload.CriticalErrors = append(rootErrorPayload.CriticalErrors, wrappedErr)
			return rootErrorPayload
		}
		rootErrorPayload.WorkspaceFetchError = wrappedErr.Error()
		return rootErrorPayload
	}
	if len(workspaces) == 0 {
		rootErrorPayload.WorkspaceFetchError = "no workspaces found in Bitbucket Cloud"
		return rootErrorPayload
	}
	bcSvc.logger.Info("Found workspaces", "count", len(workspaces))

	// var expectedWorkspace *BBktCloudWorkspace = nil
	// for _, workspace := range workspaces {
	// 	if bcSvc.config.Integrations.BitbucketCloud.Workspace == workspace.Slug {
	// 		expectedWorkspace = &workspace
	// 	}
	// }
	// if expectedWorkspace == nil {
	// 	wrappedErr := fmt.Errorf("unable to fetch %s workspace. make sure each token user is able to access the workspace: %v", bcSvc.config.Integrations.BitbucketCloud.Workspace, customerrors.ErrCritical)
	// 	bcSvc.logger.Error(wrappedErr.Error())
	// 	rootErrorPayload.CriticalErrors = append(rootErrorPayload.CriticalErrors, wrappedErr)
	// 	return rootErrorPayload
	// }

	// for _, workspace := range []BBktCloudWorkspace{*expectedWorkspace} {
	for _, workspace := range workspaces {
		workspaceError := gitdtos.BLWorkspaceError{
			WorkspaceSlug: workspace.Slug,
		}
		repos, err := bcSvc.apiClient.GetRepositoriesByWorkspace(workspace.Slug, bcSvc.dataRelayer.SendPullError)
		if err != nil {
			wrappedErr := fmt.Errorf("error pulling repositories from Bitbucket Cloud: %w", err)
			bcSvc.logger.Error(wrappedErr.Error())
			if errors.Is(err, customerrors.ErrCritical) {
				rootErrorPayload.CriticalErrors = append(rootErrorPayload.CriticalErrors, wrappedErr)
				return rootErrorPayload
			}
			workspaceError.RepoFetchError = wrappedErr.Error()
		}

		if len(repos) == 0 {
			errorMessage := fmt.Sprintf("no repositories found in workspace: %s", workspace.Slug)
			bcSvc.logger.Error(errorMessage)
			workspaceError.RepoFetchError = errorMessage
		}
		bcSvc.logger.Info("Found repositories", "count", len(repos))
		devDRepos := []gitdtos.BLRepo{}
		for _, repo := range repos {
			repoError := gitdtos.BLRepoError{
				RepoID: repo.Slug,
			}
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
			if existingRepoSyncAudit, err := bcSvc.dbQuerier.GetRepoSyncAuditByID(context.Background(), repo.Slug); err == nil {
				bcSvc.logger.Debug("Repository found in database", "name", existingRepoSyncAudit.RepoName)
				continue
			} else if errors.Is(err, sql.ErrNoRows) {
				// If the error is a no rows error, create a new repo sync audit
				bcSvc.logger.Info("Repository not found in database. Creating new repo sync audit", "name", repo.Name)
			} else {
				wrappedErr := fmt.Errorf("error getting repo sync audit for repo: %s: %w", repo.Slug, err)
				bcSvc.logger.Error(wrappedErr.Error())
				repoError.RepoProcessingError = wrappedErr.Error()
				workspaceError.RepoErrors = append(workspaceError.RepoErrors, repoError)
				continue
			}

			if _, err := bcSvc.dbQuerier.CreateRepoSyncAudit(context.Background(), dbgen.CreateRepoSyncAuditParams{
				ID:                 repo.Slug,
				RepoName:           repo.Name,
				WorkspaceSlug:      workspace.Slug,
				SuccessfulSyncTime: sql.NullTime{Valid: false},
				Success:            false,
				ErrorContext:       sql.NullString{Valid: false},
			}); err != nil {
				wrappedErr := fmt.Errorf("error creating repo sync audit for repo: %s: %w", repo.Slug, err)
				bcSvc.logger.Error(wrappedErr.Error())
				repoError.RepoProcessingError = wrappedErr.Error()
				workspaceError.RepoErrors = append(workspaceError.RepoErrors, repoError)
				continue
			}
		}

		if err := bcSvc.dataRelayer.SendCollectedData(devDRepos, url.Values(map[string][]string{"type": {"repo_pull"}})); err != nil {
			wrappedErr := fmt.Errorf("error sending pull data to data relayer: %w", err)
			bcSvc.logger.Error(wrappedErr.Error())
			if errors.Is(err, customerrors.ErrCritical) {
				rootErrorPayload.CriticalErrors = append(rootErrorPayload.CriticalErrors, wrappedErr)
				return rootErrorPayload
			}
			workspaceError.WorkspaceProcessingError = wrappedErr.Error()
		}

		if !workspaceError.IsEmpty() {
			rootErrorPayload.WorkspaceErrors = append(rootErrorPayload.WorkspaceErrors, workspaceError)
		}
	}

	if !rootErrorPayload.IsEmpty() {
		return rootErrorPayload
	}
	return nil
}

func (bcSvc *BitbucketCloudSvc) GitActivityPull() *gitdtos.BLRootErrorPayload {
	rootErrorPayload := &gitdtos.BLRootErrorPayload{}
	bcSvc.logger.Info("Pulling Git activity from Bitbucket Cloud...")
	savedRepos, err := bcSvc.getAllActiveRepoSyncAudits()
	if err != nil {
		wrappedErr := fmt.Errorf("error getting all active repo sync audits: %w", err)
		rootErrorPayload.CriticalErrors = append(rootErrorPayload.CriticalErrors, wrappedErr)
		return rootErrorPayload
	}

	bcSvc.logger.Info("Found active repo sync audits", "count", len(savedRepos))
	for _, repoSyncAudit := range savedRepos {
		bcSvc.logger.Info("Repo sync audit", "repoName", repoSyncAudit.RepoName)

		currentSyncTime := time.Now()
		if err := bcSvc.syncGitActivityForRepo(repoSyncAudit); err != nil {
			bcSvc.logger.Error("Error syncing Git activity for repo", "error", err)
			wrappedErr := fmt.Errorf("error syncing Git activity for repo: %w", err)
			bcSvc.logger.Error(wrappedErr.Error())
			if errors.Is(err, customerrors.ErrCritical) {
				rootErrorPayload.CriticalErrors = append(rootErrorPayload.CriticalErrors, wrappedErr)
				return rootErrorPayload
			}
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
				wrappedErr := fmt.Errorf("error updating repo sync audit: %w", err)
				rootErrorPayload.CriticalErrors = append(rootErrorPayload.CriticalErrors, wrappedErr)
				return rootErrorPayload
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
				wrappedErr := fmt.Errorf("error updating repo sync audit: %w", err)
				rootErrorPayload.CriticalErrors = append(rootErrorPayload.CriticalErrors, wrappedErr)
				return rootErrorPayload
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
			Limit:  int64(limit),
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

func (bcSvc *BitbucketCloudSvc) syncGitActivityForRepo(repoSyncAudit dbgen.RepositorySyncAudit) error {
	repoError := &gitdtos.BLRepoError{
		RepoID: repoSyncAudit.ID,
	}
	devDRepo := gitdtos.BLRepo{
		Slug: repoSyncAudit.ID,
	}
	var lastSuccessfulSyncTime time.Time
	if repoSyncAudit.SuccessfulSyncTime.Valid && !repoSyncAudit.SuccessfulSyncTime.Time.IsZero() {
		lastSuccessfulSyncTime = repoSyncAudit.SuccessfulSyncTime.Time
	} else {
		lastSuccessfulSyncTime = time.Now().AddDate(0, 0, -bcSvc.config.Defaults.DefaultDataPullDays)
	}
	// pull requests for the repository
	{
		fetchedPRs, err := bcSvc.apiClient.GetPullRequestsByRepository(repoSyncAudit.WorkspaceSlug, repoSyncAudit.ID, lastSuccessfulSyncTime, bcSvc.dataRelayer.SendPullError)
		if err != nil {
			wrappedErr := fmt.Errorf("error fetching pull requests for repository: %s: %w", repoSyncAudit.ID, err)
			bcSvc.logger.Error(wrappedErr.Error())
			if errors.Is(err, customerrors.ErrCritical) {
				return wrappedErr
			}
			repoError.PrFetchError = wrappedErr.Error()
		}

		devDPRs := []gitdtos.BLPullRequest{}
		for _, bBktCloudPr := range fetchedPRs {
			prError := gitdtos.BLPrError{
				PrID: bBktCloudPr.ID,
			}

			fetchedPrCommits, err := bcSvc.apiClient.GetPullRequestCommits(repoSyncAudit.WorkspaceSlug, repoSyncAudit.ID, bBktCloudPr.ID, bcSvc.dataRelayer.SendPullError)
			if err != nil {
				wrappedErr := fmt.Errorf("error fetching pull request commits for repository: %s: %w", repoSyncAudit.ID, err)
				bcSvc.logger.Error(wrappedErr.Error())
				if errors.Is(err, customerrors.ErrCritical) {
					return wrappedErr
				}
				prError.CommitFetchError = wrappedErr.Error()
			}

			devDCommits := []gitdtos.BLCommit{}
			for _, commit := range fetchedPrCommits {
				devDCommits = append(devDCommits, gitdtos.BLCommit{
					ID:                 commit.Hash,
					Message:            commit.Message,
					Committer:          convertBBktCloudUserToDevDActor(commit.Author.User, commit.Author.Raw),
					CommitterTimestamp: commit.Date,
					ChangedFiles:       []gitdtos.BLChangedFile{},
				})
			}

			isOpen := bBktCloudPr.State == string(BBktCloudPullRequestStateOpen)
			reviewers := make([]gitdtos.BLActor, len(bBktCloudPr.Reviewers))
			for i, reviewer := range bBktCloudPr.Reviewers {
				reviewers[i] = convertBBktCloudUserToDevDActor(reviewer, "")
			}
			devDPR := gitdtos.BLPullRequest{
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
				Author:       convertBBktCloudUserToDevDActor(bBktCloudPr.Author, ""),
				Reviewers:    reviewers,
				CommentCount: bBktCloudPr.CommentCount,
				Link:         bBktCloudPr.Links.HTML.Href,
				PrCommits:    devDCommits,
			}
			devDPRs = append(devDPRs, devDPR)

			if !prError.IsEmpty() {
				repoError.PrErrors = append(repoError.PrErrors, prError)
			}
		}
		if len(devDPRs) > 0 {
			devDRepo.Prs = devDPRs
		}
	}

	// commits for the repository
	{
		fetchedCommits, err := bcSvc.apiClient.GetCommitsByRepository(repoSyncAudit.WorkspaceSlug, repoSyncAudit.ID, lastSuccessfulSyncTime, bcSvc.dataRelayer.SendPullError)
		if err != nil {
			wrappedErr := fmt.Errorf("error fetching commits for repository: %s: %w", repoSyncAudit.ID, err)
			bcSvc.logger.Error(wrappedErr.Error())
			if errors.Is(err, customerrors.ErrCritical) {
				return wrappedErr
			}
			repoError.CommitFetchError = wrappedErr.Error()
		}

		devDCommits := []gitdtos.BLCommit{}
		for _, commit := range fetchedCommits {
			devDCommits = append(devDCommits, gitdtos.BLCommit{
				ID:                 commit.Hash,
				Message:            commit.Message,
				Committer:          convertBBktCloudUserToDevDActor(commit.Author.User, commit.Author.Raw),
				CommitterTimestamp: commit.Date,
				ChangedFiles:       []gitdtos.BLChangedFile{},
			})
		}
		if len(devDCommits) > 0 {
			devDRepo.Commits = devDCommits
		}
	}

	if !devDRepo.IsEmpty() {
		data := gitdtos.BLData{
			Repos: []gitdtos.BLRepo{
				devDRepo,
			},
			WorkspaceKey: repoSyncAudit.WorkspaceSlug,
		}
		if err := bcSvc.dataRelayer.SendCollectedData(data, url.Values(map[string][]string{"type": {"activity_pull"}})); err != nil {
			return fmt.Errorf("error sending data to data relayer: %w", err)
		}
	}
	if !repoError.IsEmpty() {
		if err := bcSvc.dataRelayer.SendPullError(repoError, nil); err != nil {
			return fmt.Errorf("error sending error logs to data relayer: %w", err)
		}
	}

	return nil
}

func convertBBktCloudUserToDevDActor(bBktCloudActor BBKtCloudUser, emailAddress string) gitdtos.BLActor {
	return gitdtos.BLActor{
		ID:           bBktCloudActor.AccountID,
		Name:         bBktCloudActor.DisplayName,
		DisplayName:  bBktCloudActor.DisplayName,
		EmailAddress: emailAddress,
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
