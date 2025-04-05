package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServiceKey string

const (
	BitbucketServerKey ServiceKey = "BitbucketServer"
	BitbucketCloudKey  ServiceKey = "BitbucketCloud"
	GithubKey          ServiceKey = "Github"
	JenkinsKey         ServiceKey = "Jenkins"
)

func IsValidServiceKey(key ServiceKey) bool {
	switch key {
	case BitbucketServerKey, BitbucketCloudKey, GithubKey, JenkinsKey:
		return true
	default:
		return false
	}
}

type Config struct {
	ActiveService ServiceKey   `json:"activeService"`
	Integrations  Integrations `json:"integrations"`
	Common        Common       `json:"common"`
	Defaults      Defaults     `json:"defaults"`
	Secrets       Secrets      `json:"secrets"`
}

type Integrations struct {
	BitbucketServer BitbucketServer `json:"bitbucketServer"`
	BitbucketCloud  BitbucketCloud  `json:"bitbucketCloud"`
	Github          Github          `json:"github"`
	Jenkins         Jenkins         `json:"jenkins"`
}

type BitbucketServer struct {
	URL  string `json:"url"`
	Port int    `json:"port"`
}

type BitbucketCloud struct {
	Workspace string `json:"workspace"`
}

type Github struct {
	URL  string `json:"url"`
	Port int    `json:"port"`
}

type Jenkins struct {
	URL  string `json:"url"`
	Port int    `json:"port"`
}

type Common struct {
	CronExpression      string `json:"cronExpression"`
	ReworkThresholdDays int    `json:"reworkThresholdDays"`
	OrgCode             string `json:"orgCode"`
}

type Defaults struct {
	RequestSizeThresholdInBytes      int `json:"requestSizeThresholdInBytes"`
	DefaultDataPullDays              int `json:"defaultDataPullDays"`
	WaitingTimeForRateLimitInSeconds int `json:"waitingTimeForRateLimitInSeconds"`
}

type Secrets struct {
	DDApiKey string `json:"ddApiKey"`
}

func NewConfig(filePath string) (*Config, error) {
	config := &Config{
		Defaults: Defaults{
			RequestSizeThresholdInBytes:      200 * 1024, // 200KB
			DefaultDataPullDays:              30,
			WaitingTimeForRateLimitInSeconds: 3600,
		},
	}
	// Load configuration from JSON file or environment variables
	// For example, you can use encoding/json to unmarshal a JSON file into the config struct
	// Or use os.Getenv to load environment variables into the config struct
	// Here, we will just load from a file for simplicity
	// You can also use a library like viper for more advanced configuration management
	err := loadConfigFromFile(filePath, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func loadConfigFromFile(filePath string, config *Config) error {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, config)
}

func (c *Config) ValidateDefaultsAndCommonConfig() error {
	// Add validation logic here
	if c.ActiveService == "" {
		return fmt.Errorf("activeService is required")
	}

	if c.Common.CronExpression == "" {
		return fmt.Errorf("cronExpression is required")
	}
	if c.Common.ReworkThresholdDays <= 0 {
		return fmt.Errorf("reworkThresholdDays must be greater than 0")
	}
	if c.Defaults.RequestSizeThresholdInBytes <= 0 || c.Defaults.RequestSizeThresholdInBytes >= 200*1024 {
		// AWS SQS max message size is 256KB. keeping 200KB as threshold and 56 KB for overhead buffer
		return fmt.Errorf("requestSizeThresholdInBytes must be between 0KB and 200KB")
	}
	if c.Defaults.DefaultDataPullDays <= 1 {
		return fmt.Errorf("defaultDataPullDays must be greater than 1")
	}
	if c.Defaults.WaitingTimeForRateLimitInSeconds <= 0 {
		return fmt.Errorf("waitingTimeForRateLimitInSeconds must be greater than 0")
	}
	return nil
}

func (c *Config) GetServiceConfig() (interface{}, error) {
	switch c.ActiveService {
	case BitbucketServerKey:
		return c.Integrations.BitbucketServer, nil
	case BitbucketCloudKey:
		return c.Integrations.BitbucketCloud, nil
	case GithubKey:
		return c.Integrations.Github, nil
	case JenkinsKey:
		return c.Integrations.Jenkins, nil
	default:
		return nil, fmt.Errorf("unsupported service key: %s", c.ActiveService)
	}
}
