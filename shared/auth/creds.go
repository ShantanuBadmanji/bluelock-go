package auth

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	credKey  string
}

func NewCredentials(username, password string) *Credential {
	return &Credential{
		Username: username,
		Password: password,
	}
}

func (c *Credential) GetCredKey() string {
	return c.credKey
}

func (c *Credential) GetCredential() (string, string) {
	return c.Username, c.Password
}

func (c *Credential) GenerateCredKey() string {
	if c.credKey == "" {
		c.credKey = c.Username + ":" + c.Password
	}
	return c.credKey
}
