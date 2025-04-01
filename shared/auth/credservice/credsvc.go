package credservice

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bluelock-go/shared/auth"
)

type AuthTokens struct {
	DatapullCredentials       []auth.Credential `json:"datapullCredentials"`
	CommitAnalysisCredentials []auth.Credential `json:"commitAnalysisCredentials"`
}

func LoadAuthTokensFromFile(filePath string) (*AuthTokens, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var tokens AuthTokens
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate the loaded tokens
	// Validate datapull credentials
	if tokens.DatapullCredentials == nil {
		return nil, fmt.Errorf("datapull credentials are required")
	}
	for _, cred := range tokens.DatapullCredentials {
		if cred.GetUsername() == "" || cred.GetPassword() == "" {
			return nil, fmt.Errorf("invalid datapull credentials: username and password must not be empty")
		}
	}

	// Validate commit analysis credentials if they exist
	if tokens.CommitAnalysisCredentials != nil {
		for _, cred := range tokens.CommitAnalysisCredentials {
			if cred.GetUsername() == "" || cred.GetPassword() == "" {
				return nil, fmt.Errorf("invalid commit analysis credentials: username and password must not be empty")
			}
		}
	}

	return &tokens, nil
}
