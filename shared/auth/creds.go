package auth

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	CredKey  string
}

func NewCredentials(username, password string) *Credential {
	return &Credential{
		Username: username,
		Password: password,
	}
}

func (c *Credential) GetUsername() string {
	return c.Username
}

func (c *Credential) GetPassword() string {
	return c.Password
}

func (c *Credential) GetCredKey() string {
	return c.CredKey
}

func (c *Credential) GetCredential() (string, string) {
	return c.Username, c.Password
}

func (c *Credential) GenerateCredKey() string {
	if c.CredKey == "" {
		c.CredKey = c.Username + ":" + c.Password
	}
	return c.CredKey
}
