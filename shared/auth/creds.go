package auth

import "encoding/base64"

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
