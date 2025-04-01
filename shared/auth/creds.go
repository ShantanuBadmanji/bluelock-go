package auth

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

func (c *Credential) GetCredential() (string, string) {
	return c.Username, c.Password
}

