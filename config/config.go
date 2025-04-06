package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bluelock-go/shared"
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

func NewDefaults() *Defaults {
	return &Defaults{
		RequestSizeThresholdInBytes:      200 * 1024, // 200KB
		DefaultDataPullDays:              30,
		WaitingTimeForRateLimitInSeconds: 3600,
	}
}

func NewConfig(filePath string) (*Config, error) {
	config := &Config{}
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

func LoadMergedConfig() (*Config, error) {
	defaultConfigFilePath := filepath.Join(shared.RootDir, "config", "config.json")

	defaultConfig, err := NewConfig(defaultConfigFilePath)
	if err != nil {
		return nil, err
	}

	// Validate the loaded configuration
	err = defaultConfig.ValidateDefaultsAndCommonConfig()
	if err != nil {
		return nil, err
	}

	userConfigFilePath := filepath.Join(shared.RootDir, "config", "config.user.json")
	userConfig, err := NewConfig(userConfigFilePath)
	if err != nil {
		return nil, err
	}

	// Validate the loaded user configuration
	isValidServiceKey := IsValidServiceKey(userConfig.ActiveService)
	if !isValidServiceKey {
		return nil, fmt.Errorf("invalid activeService key: %s", userConfig.ActiveService)
	}

	// Merge user configuration with default configuration
	mergedConfig := &Config{}

	// Copy default config values
	*mergedConfig = *defaultConfig
	mergedConfig.ActiveService = userConfig.ActiveService
	switch mergedConfig.ActiveService {
	case BitbucketServerKey:
		if userConfig.Integrations.BitbucketServer.URL == "" {
			return nil, fmt.Errorf("bitbucketServer URL is required")
		}
		mergedConfig.Integrations.BitbucketServer = userConfig.Integrations.BitbucketServer
	case BitbucketCloudKey:
		if userConfig.Integrations.BitbucketCloud.Workspace == "" {
			return nil, fmt.Errorf("bitbucketCloud Workspace is required")
		}
		mergedConfig.Integrations.BitbucketCloud = userConfig.Integrations.BitbucketCloud
	case GithubKey:
		if userConfig.Integrations.Github.URL != defaultConfig.Integrations.Github.URL {
			mergedConfig.Integrations.Github = userConfig.Integrations.Github
		}
	case JenkinsKey:
		if userConfig.Integrations.Jenkins.URL == "" {
			return nil, fmt.Errorf("jenkins URL is required")
		}
		mergedConfig.Integrations.Jenkins = userConfig.Integrations.Jenkins
	default:
		return nil, fmt.Errorf("unsupported service key: %s", userConfig.ActiveService)
	}

	// Merge common values
	if userConfig.Common.OrgCode == "" {
		return nil, fmt.Errorf("orgCode is required")
	} else {
		mergedConfig.Common.OrgCode = defaultConfig.Common.OrgCode
	}

	if userConfig.Common.CronExpression != "" {
		mergedConfig.Common.CronExpression = userConfig.Common.CronExpression
	}
	if userConfig.Common.ReworkThresholdDays != 0 {
		mergedConfig.Common.ReworkThresholdDays = userConfig.Common.ReworkThresholdDays
	}

	// Merge default values
	if userConfig.Defaults.RequestSizeThresholdInBytes != 0 {
		mergedConfig.Defaults.RequestSizeThresholdInBytes = userConfig.Defaults.RequestSizeThresholdInBytes
	}
	if userConfig.Defaults.DefaultDataPullDays != 0 {
		mergedConfig.Defaults.DefaultDataPullDays = userConfig.Defaults.DefaultDataPullDays
	}
	if userConfig.Defaults.WaitingTimeForRateLimitInSeconds != 0 {
		mergedConfig.Defaults.WaitingTimeForRateLimitInSeconds = userConfig.Defaults.WaitingTimeForRateLimitInSeconds
	}
	if userConfig.Secrets.DDApiKey != "" {
		mergedConfig.Secrets.DDApiKey = userConfig.Secrets.DDApiKey
	}
	// Validate the merged configuration
	err = mergedConfig.ValidateDefaultsAndCommonConfig()
	if err != nil {
		return nil, err
	}

	return mergedConfig, nil
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
