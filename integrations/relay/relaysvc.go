package relay

import (
	"fmt"

	"github.com/bluelock-go/config"
)

type DataRelayService interface {
	SendCollectedData(payload interface{}) error
	SendPullError(payload interface{}) error
	SendDataAndError(payload interface{}, err error) error
}

const (
	RelayBaseURL = "https://relay.bluelock.com"
)

type DevDynamicsRelayService struct {
	APIKey string
}

func NewDevDynamicsRelayService(apiKey string) *DevDynamicsRelayService {
	return &DevDynamicsRelayService{
		APIKey: apiKey,
	}
}

func (d *DevDynamicsRelayService) SendCollectedData(payload interface{}) error {
	// Implement the logic to send collected data to the relay service
	// This is a placeholder implementation
	return nil
}

func (d *DevDynamicsRelayService) SendPullError(payload interface{}) error {
	// Implement the logic to send pull error to the relay service
	// This is a placeholder implementation
	return nil
}
func (d *DevDynamicsRelayService) SendDataAndError(payload interface{}, err error) error {
	// Implement the logic to send both data and error to the relay service
	// This is a placeholder implementation
	return nil
}

func GetAPIKeyFromConfig(config *config.Config) (string, error) {
	apiKey := config.Secrets.DDApiKey
	if apiKey == "" {
		return "", fmt.Errorf("API key not found in the configuration")
	}
	return apiKey, nil
}
