package auth

import (
	"encoding/base64"
	"fmt"

	"github.com/bluelock-go/shared/customerrors"
)

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	CredKey  string `json:"credKey"`
}

func NewCredentials(username, password string) *Credential {
	return &Credential{
		Username: username,
		Password: password,
	}
}

func (c *Credential) GetCredential() (string, string) {
	return c.Username, c.Password
}

func (c *Credential) GenerateCredKeyIfAbsent() string {
	if c.CredKey == "" {
		c.CredKey = base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))
	}
	return c.CredKey
}

func GetCredentialByCredKey(credKey string, creds []Credential) (*Credential, error) {
	for _, cred := range creds {
		if cred.CredKey == credKey {
			return &cred, nil
		}
	}
	return nil, fmt.Errorf("credential with credKey %s not found: %w", credKey, customerrors.ErrCritical)
}

func ValidateCredentials(credStoreKey string, creds []Credential) error {
	for _, cred := range creds {
		if cred.CredKey == "" {
			return fmt.Errorf("invalid credentials: CredKey must not be empty for authCredentialStoreKey %s", credStoreKey)
		}
		if cred.Username == "" || cred.Password == "" {
			return fmt.Errorf("invalid credentials for key %s: Username and Password must not be empty", cred.CredKey)
		}
	}
	return nil
}
