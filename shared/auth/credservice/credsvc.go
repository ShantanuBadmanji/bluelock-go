package credservice

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bluelock-go/shared/auth"
)

type AuthCredentialStore map[string][]auth.Credential

const (
	DatapullCredentialsKey       = "datapullCredentials"
	CommitAnalysisCredentialsKey = "commitAnalysisCredentials"
)

func LoadAuthTokensFromFile(filePath string) (AuthCredentialStore, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var credStore AuthCredentialStore
	if err := json.Unmarshal(data, &credStore); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate the loaded credentials
	if credStore == nil {
		return nil, fmt.Errorf("credential store is nil")
	}
	if len(credStore) == 0 {
		return nil, fmt.Errorf("credential store is empty")
	}
	if _, ok := credStore[DatapullCredentialsKey]; !ok {
		return nil, fmt.Errorf("missing required key: %s", DatapullCredentialsKey)
	}

	for key, creds := range credStore {
		for _, cred := range creds {
			if cred.Username == "" || cred.Password == "" {
				return nil, fmt.Errorf("invalid credentials for key %s: username and password must not be empty", key)
			}
		}
	}

	// populate the CredKey for each credential
	for key, creds := range credStore {
		for i := range creds {
			creds[i].GenerateCredKey()
		}
		credStore[key] = creds
	}

	return credStore, nil
}
