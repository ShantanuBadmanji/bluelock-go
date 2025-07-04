package credservice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/bluelock-go/shared/auth"
	"github.com/gofrs/flock"
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

	//  write in a temporary file and rename it to the original file
	if err := atomicWriteFile(filePath, credStoreInBytes); err != nil {
		return nil, fmt.Errorf("failed to write updated credentials to file as atomic operation failed: %w", err)
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
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	lock := flock.New(filePath + ".lock")
	if ok, err := lock.TryLockContext(ctx, 10*time.Millisecond); err != nil {
		return nil, nil, fmt.Errorf("failed to acquire lock: %w", err)
	} else if !ok {
		return nil, nil, fmt.Errorf("failed to acquire lock in 100 milliseconds: another process is holding the lock")
	}

	// immediately unlock the lock after acquiring it
	// to avoid deadlock in case of long running operations or making the other process wait
	// Since we are using a read lock, we can release it immediately
	// and let the other process acquire it
	lock.Unlock()

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
	re := regexp.MustCompile(`\.json$`)
	backupFilePath := re.ReplaceAllString(filePath, ".backup.json")
	if err := os.WriteFile(backupFilePath, credStoreInBytes, 0644); err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}

	return nil
}

func atomicWriteFile(filePath string, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lock := flock.New(filePath + ".lock")
	if ok, err := lock.TryLockContext(ctx, 10*time.Millisecond); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	} else if !ok {
		return fmt.Errorf("failed to acquire lock in 5 seconds: another process is holding the lock")
	}
	defer lock.Unlock()

	// Create a temporary file
	tempFilePath := filePath + ".tmp"
	if err := os.WriteFile(tempFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Rename the temporary file to the original file
	if err := os.Rename(tempFilePath, filePath); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
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
		if err := auth.ValidateCredentials(key, creds); err != nil {
			return err
		}
		for _, cred := range creds {
			if cred.Username == "" || cred.Password == "" {
				return fmt.Errorf("invalid credentials for key %s: username and password must not be empty", key)
			}
		}
	}

	return nil
}
