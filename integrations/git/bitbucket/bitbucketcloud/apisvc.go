package bitbucketcloud

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
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

func NewClient(apiProvider di.APIProvider) *Client {
	return &Client{
		baseURL:      "https://api.bitbucket.org/2.0",
		httpClient:   apiProvider.GetHTTPClient(),
		stateManager: apiProvider.GetStateManager(),
		logger:       apiProvider.GetLogger(),
		credentials:  apiProvider.GetCredentials(),
	}
}

const MAX_ATTEMPTS = 2
const WAITING_TIME_FOR_RATE_LIMIT_IN_SECONDS = 3

func (c *Client) HandleRequestWithRetries(requestCallback func(*auth.Credential) (*http.Response, error)) (*http.Response, error) {
	for attemptNumber := range MAX_ATTEMPTS {
		c.logger.Info(fmt.Sprintf("Attempt number: %d", attemptNumber+1))

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
			if 200 <= response.StatusCode && response.StatusCode < 300 {
				if response.StatusCode == 200 {
					return response, nil
				}

				// Handle 2xx responses other than 200
				c.logger.Error("Successful response received for token: " + authCred.CredKey)

				defer response.Body.Close()
				var message string
				if body, err := io.ReadAll(response.Body); err != nil {
					c.logger.Error("Failed to read response body: " + err.Error())
					message = fmt.Sprintf("failed to read response body: %s", err.Error())
				} else {
					message = fmt.Sprintf("response body: %s", string(body))
				}

				return nil, fmt.Errorf("unexpected 2xx response code: %d for token: %s. message: %s", response.StatusCode, authCred.CredKey, message)

			} else {
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
					c.logger.Error(fmt.Sprintf("Unhandled response code: %d for token: %s", response.StatusCode, authCred.CredKey))
					return nil, fmt.Errorf("unhandled response code: %d for token: %s", response.StatusCode, authCred.CredKey)
				}
			}
		}
	}

	return nil, fmt.Errorf("exceeded maximum reset limit(%d) without a successful response", MAX_ATTEMPTS)
}

func (c *Client) getRequestCallback(url string, sendErrorLogCallback func(string)) func(*auth.Credential) (*http.Response, error) {

	return func(cred *auth.Credential) (*http.Response, error) {
		token := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, fmt.Sprintf("%s:%s", cred.Username, cred.Password)))
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			logMessage := fmt.Sprintf("Failed to create new request: %s", err.Error())
			c.logger.Error(logMessage)
			sendErrorLogCallback(logMessage)
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", token))

		return c.httpClient.Do(req)
	}
}

func (c *Client) GetWorkspaces(sendErrorLogCallback func(string)) ([]BBktCloudWorkspace, error) {
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
			c.logger.Info("No more pages to fetch for workspaces.")
			break
		}
	}

	return workspaces, nil
}

func (c *Client) GetRepositoriesByWorkspace(workspace string, sendErrorLogCallback func(string)) ([]BBktCloudRepository, error) {
	repositories := []BBktCloudRepository{}
	pageLen := 100

	url := fmt.Sprintf("%s/repositories/%s?pagelen=%d", c.baseURL, workspace, pageLen)

	for len(url) > 0 {
		response, err := c.HandleRequestWithRetries(c.getRequestCallback(url, sendErrorLogCallback))
		if err != nil {
			logMessage := fmt.Sprintf("Failed to get repositories for workspace %s: %s", workspace, err.Error())
			c.logger.Error(logMessage)
			return nil, fmt.Errorf("failed to get repositories for workspace %s: %w", workspace, err)
		}
		defer response.Body.Close()

		var repoResponse BBktCloudPaginatedResponse[BBktCloudRepository]
		if err := json.NewDecoder(response.Body).Decode(&repoResponse); err != nil {
			logMessage := fmt.Sprintf("Failed to decode repositories response for workspace %s: %s", workspace, err.Error())
			c.logger.Error(logMessage)
			return repositories, fmt.Errorf("failed to decode repositories response for workspace %s: %w", workspace, err)
		}

		repositories = append(repositories, repoResponse.Values...)

		url = repoResponse.Next
		if url == "" {
			c.logger.Info(fmt.Sprintf("No more pages to fetch for repositories in workspace %s.", workspace))
			break
		}
	}

	return repositories, nil
}

func (c *Client) GetPullRequestsByRepository(workspace, repository string, sendErrorLogCallback func(string)) ([]BBktCloudPullRequest, error) {
	pullRequests := []BBktCloudPullRequest{}
	pageLen := 100

	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests?q=state IN (\"OPEN\", \"MERGED\", \"DECLINED\", \"SUPERSEDED\") AND updated_on >= {iso8601_last_run_time}&pagelen=%d", c.baseURL, workspace, repository, pageLen)

	for len(url) > 0 {
		response, err := c.HandleRequestWithRetries(c.getRequestCallback(url, sendErrorLogCallback))
		if err != nil {
			logMessage := fmt.Sprintf("Failed to get pull requests for repository %s/%s: %s", workspace, repository, err.Error())
			c.logger.Error(logMessage)
			return nil, fmt.Errorf("failed to get pull requests for repository %s/%s: %w", workspace, repository, err)
		}

		defer response.Body.Close()

		var prResponse BBktCloudPaginatedResponse[BBktCloudPullRequest]
		if err := json.NewDecoder(response.Body).Decode(&prResponse); err != nil {
			logMessage := fmt.Sprintf("Failed to decode pull requests response for repository %s/%s: %s", workspace, repository, err.Error())
			c.logger.Error(logMessage)
			return pullRequests, fmt.Errorf("failed to decode pull requests response for repository %s/%s: %w", workspace, repository, err)
		}

		pullRequests = append(pullRequests, prResponse.Values...)

		url = prResponse.Next
		if url == "" {
			c.logger.Info(fmt.Sprintf("No more pages to fetch for pull requests in repository %s/%s.", workspace, repository))
			break
		}
	}

	return pullRequests, nil
}
