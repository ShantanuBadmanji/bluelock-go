package credservice

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/bluelock-go/shared/auth"
)

type AuthCredentialStore map[string][]auth.Credential

const (
	DatapullCredentialsKey       = "datapullCredentials"
	CommitAnalysisCredentialsKey = "commitAnalysisCredentials"
)

func NormalizeAndPersistCredentials(filePath string) (AuthCredentialStore, error) {
	credStore, data, err := LoadAuthTokensFromFileAndValidate(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load and validate credentials: %w", err)
	}

	// Create a backup file path by replacing the original file name with a backup name
	if err := takeBackupOfCredStore(filePath, data); err != nil {
		return nil, fmt.Errorf("failed to create backup of credential store: %w", err)
	}

	// populate the CredKey for each credential
	for key, creds := range credStore {
		for i := range creds {
			creds[i].GenerateCredKeyIfAbsent()
		}
		credStore[key] = creds
	}

	// Write the updated credentials back to the original file
	credStoreInBytes, err := json.MarshalIndent(credStore, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated credentials: %w", err)
	}
	if err := os.WriteFile(filePath, credStoreInBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write updated credentials to file: %w", err)
	}

	// Load and validate the updated credentials
	// This is to ensure that the updated credentials are valid after writing them back to the file
	// and to avoid any potential issues with the file format
	if credStore, _, err = LoadAuthTokensFromFileAndValidate(filePath); err != nil {
		return nil, fmt.Errorf("failed to load and validate updated credentials: %w", err)
	}

	return credStore, nil
}

func LoadAuthTokensFromFileAndValidate(filePath string) (AuthCredentialStore, []byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	var credStore AuthCredentialStore
	if err := json.Unmarshal(data, &credStore); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if err := credStore.validateCredStore(); err != nil {
		return nil, nil, fmt.Errorf("invalid credential store: %w", err)
	}

	return credStore, data, nil
}

func takeBackupOfCredStore(filePath string, credStoreInBytes []byte) error {
	backupFilePath := filePath
	re := regexp.MustCompile(`\.json$`)
	backupFilePath = re.ReplaceAllString(filePath, ".backup.json")
	if err := os.WriteFile(backupFilePath, credStoreInBytes, 0644); err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}

	return nil
}

func (credStore AuthCredentialStore) validateCredStore() error {
	if credStore == nil {
		return fmt.Errorf("credential store is nil")
	}
	if len(credStore) == 0 {
		return fmt.Errorf("credential store is empty")
	}
	if _, ok := credStore[DatapullCredentialsKey]; !ok {
		return fmt.Errorf("missing required key: %s", DatapullCredentialsKey)
	}

	for key, creds := range credStore {
		for _, cred := range creds {
			if cred.Username == "" || cred.Password == "" {
				return fmt.Errorf("invalid credentials for key %s: username and password must not be empty", key)
			}
		}
	}

	return nil
}
