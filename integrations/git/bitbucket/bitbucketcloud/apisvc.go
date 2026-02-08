package bitbucketcloud

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
	"github.com/bluelock-go/shared/auth/credservice"
	"github.com/bluelock-go/shared/customerrors"
	"github.com/bluelock-go/shared/di"
	"github.com/bluelock-go/shared/storage/state/statemanager"
)

type Client struct {
	baseURL      string
	httpClient   *http.Client
	stateManager *statemanager.StateManager
	logger       *shared.CustomLogger
	credentials  []auth.Credential
}

func NewClient(httpClient *http.Client, stateManager *statemanager.StateManager, logger *shared.CustomLogger, credentials []auth.Credential) *Client {
	return &Client{
		baseURL:      "https://api.bitbucket.org/2.0",
		httpClient:   httpClient,
		stateManager: stateManager,
		logger:       logger,
		credentials:  credentials,
	}
}

const MAX_ATTEMPTS = 2
const WAITING_TIME_FOR_RATE_LIMIT_IN_SECONDS = 3

func (c *Client) HandleRequestWithRetries(requestCallback func(*auth.Credential) (*http.Response, error)) (*http.Response, error) {
	for attemptNumber := range MAX_ATTEMPTS {

		if attemptNumber > 0 {
			c.logger.Info(fmt.Sprintf("Sleeping for %d seconds", WAITING_TIME_FOR_RATE_LIMIT_IN_SECONDS))
			time.Sleep(WAITING_TIME_FOR_RATE_LIMIT_IN_SECONDS * time.Second)
			c.logger.Info("Woke up!!\nResetting usage metrics for all tokens")
			c.stateManager.ResetUsageMetricsForAllTokens(time.Now())
			c.logger.Info("Retrying...")
		}

		for {
			activeTokenID, err := c.stateManager.GetLeastUsageActiveToken()
			if err != nil {
				c.logger.Error("Failed to get least usage active token: " + err.Error())
				c.logger.Warn("Current token states: ", "tokenStates", c.stateManager.State.TokenStates)
				if errors.Is(err, customerrors.ErrCritical) {
					return nil, err
				} else if errors.Is(err, statemanager.ErrAllTokensExhausted) {
					c.logger.Warn("All tokens are exhausted, need to wait for rate limit to reset.")
					break
				}
			}

			if activeTokenID == "" {
				c.logger.Error("Token ID is empty but no error was returned")
				return nil, fmt.Errorf("activeTokenID is empty: %w", customerrors.ErrCritical)
			}

			authCred, err := auth.GetCredentialByCredKey(activeTokenID, c.credentials)
			if err != nil {
				c.logger.Error("Failed to get credential by credKey: " + err.Error())
				break
			}
			if authCred == nil {
				c.logger.Error("Credential is nil but no error was returned")
				// set the cread to unauthorized
				c.stateManager.SetTokenStatusToUnauthorized(activeTokenID)
				c.logger.Warn("Retrying with next available token.")
				continue
			}

			response, err := requestCallback(authCred)
			if err != nil {
				return nil, err
			}
			if response.StatusCode == 200 {
				return response, nil
			}
			defer response.Body.Close()
			var message string
			if body, err := io.ReadAll(response.Body); err != nil {
				c.logger.Error("Failed to read response body: " + err.Error())
				message = fmt.Sprintf("failed to read response body: %s", err.Error())
			} else {
				message = fmt.Sprintf("response body: %s", string(body))
			}

			switch response.StatusCode {
			case 401:
				// Handle Unauthorized
				c.logger.Error("Unauthorized access for token: " + authCred.CredKey)
				c.stateManager.SetTokenStatusToUnauthorized(authCred.CredKey)
			case 429:
				// Handle Rate Limit Exceeded
				c.logger.Warn("Rate limit exceeded for token: " + authCred.CredKey)
				c.stateManager.SetTokenStatusToRateLimited(authCred.CredKey)
			default:
				c.logger.Error(fmt.Sprintf("Unhandled response code: %d for token: %s. message: %s", response.StatusCode, authCred.CredKey, message))
				return nil, fmt.Errorf("unhandled response code: %d for token: %s. message: %s", response.StatusCode, authCred.CredKey, message)
			}
		}
	}

	return nil, fmt.Errorf("exceeded maximum reset limit(%d) without a successful response", MAX_ATTEMPTS)
}

func (c *Client) getRequestCallback(url string, sendErrorLogCallback func(payload interface{}, queryParams url.Values) error) func(*auth.Credential) (*http.Response, error) {

	return func(cred *auth.Credential) (*http.Response, error) {
		token := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, fmt.Sprintf("%s:%s", cred.Username, cred.Password)))
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			wrappedErr := fmt.Errorf("Failed to create new request: %w", err)
			c.logger.Error(wrappedErr.Error())
			sendErrorLogCallback(wrappedErr.Error(), nil)
			return nil, wrappedErr
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", token))

		response, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
		if err != nil {
			wrappedErr := fmt.Errorf("failed to execute request: %w", err)
			c.logger.Error(wrappedErr.Error())
			sendErrorLogCallback(wrappedErr.Error(), nil)
			return nil, wrappedErr
		}
		return response, nil
	}
}

func (c *Client) GetWorkspaces(sendErrorLogCallback func(payload interface{}, queryParams url.Values) error) ([]BBktCloudWorkspace, error) {
	workspaces := []BBktCloudWorkspace{}
	pageLen := 50

	url := fmt.Sprintf("%s/workspaces?pagelen=%d", c.baseURL, pageLen)

	for len(url) > 0 {
		response, err := c.HandleRequestWithRetries(c.getRequestCallback(url, sendErrorLogCallback))
		if err != nil {
			logMessage := fmt.Sprintf("Failed to get workspaces: %s", err.Error())
			c.logger.Error(logMessage)
			return nil, fmt.Errorf("failed to get workspaces: %w", err)
		}
		defer response.Body.Close()

		var workspaceResponse BBktCloudPaginatedResponse[BBktCloudWorkspace]
		if err := json.NewDecoder(response.Body).Decode(&workspaceResponse); err != nil {
			logMessage := fmt.Sprintf("Failed to decode workspaces response: %s", err.Error())
			c.logger.Error(logMessage)
			return workspaces, fmt.Errorf("failed to decode workspaces response: %w", err)
		}

		workspaces = append(workspaces, workspaceResponse.Values...)

		url = workspaceResponse.Next
		if url == "" {
			break
		}
	}

	return workspaces, nil
}

func (c *Client) GetRepositoriesByWorkspace(workspace string, sendErrorLogCallback func(payload interface{}, queryParams url.Values) error) ([]BBktCloudRepository, error) {
	repositories := []BBktCloudRepository{}
	pageLen := 100

	url := fmt.Sprintf("%s/repositories/%s?pagelen=%d", c.baseURL, workspace, pageLen)

	for len(url) > 0 {
		response, err := c.HandleRequestWithRetries(c.getRequestCallback(url, sendErrorLogCallback))
		if err != nil {
			logMessage := fmt.Sprintf("Failed to get repositories for workspace url: %s: %s", url, err.Error())
			c.logger.Error(logMessage)
			return nil, fmt.Errorf("failed to get repositories for workspace url: %s: %w", url, err)
		}
		defer response.Body.Close()

		var repoResponse BBktCloudPaginatedResponse[BBktCloudRepository]
		if err := json.NewDecoder(response.Body).Decode(&repoResponse); err != nil {
			logMessage := fmt.Sprintf("Failed to decode repositories response for workspace url: %s: %s", url, err.Error())
			c.logger.Error(logMessage)
			return repositories, fmt.Errorf("failed to decode repositories response for workspace url: %s: %w", url, err)
		}

		repositories = append(repositories, repoResponse.Values...)

		url = repoResponse.Next
		if url == "" {
			break
		}
	}

	return repositories, nil
}

func (c *Client) GetPullRequestsByRepository(workspace, repository string, lastSuccessfulSyncTime time.Time, sendErrorLogCallback func(payload interface{}, queryParams url.Values) error) ([]BBktCloudPullRequest, error) {
	pullRequests := []BBktCloudPullRequest{}
	pageLen := 50

	lastSuccessfulSyncTimeUTCString := lastSuccessfulSyncTime.UTC().Format(time.RFC3339)
	urlQueryParams := url.Values{}
	urlQueryParams.Add("q", fmt.Sprintf("state IN (\"OPEN\", \"MERGED\", \"DECLINED\", \"SUPERSEDED\") AND updated_on >= %s", lastSuccessfulSyncTimeUTCString))
	urlQueryParams.Add("pagelen", fmt.Sprintf("%d", pageLen))
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests?%s", c.baseURL, workspace, repository, urlQueryParams.Encode())

	for len(url) > 0 {
		response, err := c.HandleRequestWithRetries(c.getRequestCallback(url, sendErrorLogCallback))
		if err != nil {
			logMessage := fmt.Sprintf("Failed to get pull requests for repository url: %s: %s", url, err.Error())
			c.logger.Error(logMessage)
			return nil, fmt.Errorf("failed to get pull requests for repository url: %s: %w", url, err)
		}
		defer response.Body.Close()
		var prResponse BBktCloudPaginatedResponse[BBktCloudPullRequest]
		if err := json.NewDecoder(response.Body).Decode(&prResponse); err != nil {
			logMessage := fmt.Sprintf("Failed to decode pull requests response for repository url: %s: %s", url, err.Error())
			c.logger.Error(logMessage)
			return pullRequests, fmt.Errorf("failed to decode pull requests response for repository url: %s: %w", url, err)
		}

		pullRequests = append(pullRequests, prResponse.Values...)

		url = prResponse.Next
		if url == "" {
			break
		}
	}

	return pullRequests, nil
}

func (c *Client) GetPullRequestCommits(workspace, repository string, pullRequestID int, sendErrorLogCallback func(payload interface{}, queryParams url.Values) error) ([]BBktCloudCommit, error) {
	commits := []BBktCloudCommit{}
	pageLen := 100

	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d/commits?pagelen=%d", c.baseURL, workspace, repository, pullRequestID, pageLen)

	for len(url) > 0 {
		response, err := c.HandleRequestWithRetries(c.getRequestCallback(url, sendErrorLogCallback))
		if err != nil {
			return nil, fmt.Errorf("failed to get pull request commits for repository url: %s: %w", url, err)
		}

		defer response.Body.Close()

		var commitResponse BBktCloudPaginatedResponse[BBktCloudCommit]
		if err := json.NewDecoder(response.Body).Decode(&commitResponse); err != nil {
			return commits, fmt.Errorf("failed to decode commits response for repository url: %s: %w", url, err)
		}

		commits = append(commits, commitResponse.Values...)

		url = commitResponse.Next
		if url == "" {
			break
		}
	}

	return commits, nil
}

func (c *Client) GetCommitsByRepository(workspace, repository string, lastSuccessfulSyncTime time.Time, sendErrorLogCallback func(payload interface{}, queryParams url.Values) error) ([]BBktCloudCommit, error) {
	commits := []BBktCloudCommit{}
	pageLen := 100

	url := fmt.Sprintf("%s/repositories/%s/%s/commits?pagelen=%d", c.baseURL, workspace, repository, pageLen)

	for len(url) > 0 {
		response, err := c.HandleRequestWithRetries(c.getRequestCallback(url, sendErrorLogCallback))
		if err != nil {
			return nil, fmt.Errorf("failed to get commits for repository url: %s: %w", url, err)
		}

		defer response.Body.Close()

		var commitResponse BBktCloudPaginatedResponse[BBktCloudCommit]
		if err := json.NewDecoder(response.Body).Decode(&commitResponse); err != nil {
			return commits, fmt.Errorf("failed to decode commits response for repository url: %s: %w", url, err)
		}

		filteredCommits := []BBktCloudCommit{}
		for _, commit := range commitResponse.Values {
			if commit.Date.After(lastSuccessfulSyncTime) {
				filteredCommits = append(filteredCommits, commit)
			}
		}
		commits = append(commits, filteredCommits...)

		if len(filteredCommits) < pageLen {
			break
		}

		url = commitResponse.Next
		if url == "" {
			break
		}
	}

	return commits, nil
}

var client = di.NewThreadSafeSingleton(func() *Client {
	customLogger := shared.AcquireCustomLogger()
	stateManager := statemanager.AcquireStateManager()
	credentials := credservice.AcquireCredentials()
	return NewClient(http.DefaultClient, stateManager, customLogger, credentials)
})

func AcquireClient() *Client {
	return client.Acquire()
}
