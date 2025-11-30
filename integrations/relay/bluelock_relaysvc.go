package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/shared/di"
)

type BluelockRelayService struct {
	BaseURL string
	APIKey  string
}

func NewBluelockRelayService(relayBaseURL string, orgCode string, activeIntegrationService config.ServiceKey, apiKey string) *BluelockRelayService {
	baseURL := fmt.Sprintf("%s/api/v1/bluelock/%s/%s", relayBaseURL, orgCode, activeIntegrationService)
	return &BluelockRelayService{baseURL, apiKey}
}

func (blrsvc *BluelockRelayService) SendCollectedData(payload interface{}, queryParams url.Values) error {
	dataPayload := map[string]interface{}{
		"data": payload,
	}
	jsonPayload, err := json.Marshal(dataPayload)
	if err != nil {
		return fmt.Errorf("failed to send collected data: error marshalling data payload: %w", err)
	}

	url := fmt.Sprintf("%s/pull-data", blrsvc.BaseURL)
	if queryParams != nil {
		url = fmt.Sprintf("%s?%s", url, queryParams.Encode())
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send collected data: error creating request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+blrsvc.APIKey)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send collected data: error making request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send collected data: status code %d", response.StatusCode)
	}
	return nil
}

func (blrsvc *BluelockRelayService) SendPullError(payload interface{}, queryParams url.Values) error {
	errorPayload := map[string]interface{}{
		"error": payload,
	}
	jsonPayload, err := json.Marshal(errorPayload)
	if err != nil {
		return fmt.Errorf("failed to send pull error: error marshalling error payload: %w", err)
	}

	url := fmt.Sprintf("%s/pull-error", blrsvc.BaseURL)
	if queryParams != nil {
		url = fmt.Sprintf("%s?%s", url, queryParams.Encode())
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send pull error: error creating request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+blrsvc.APIKey)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send pull error: error making request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send pull error: status code %d", response.StatusCode)
	}

	return nil
}

func (blrsvc *BluelockRelayService) SendDataAndError(dataPayload interface{}, errorPayload interface{}, queryParams url.Values) error {
	if dataPayload == nil {
		return fmt.Errorf("data payload is nil")
	}
	dataErr := blrsvc.SendCollectedData(dataPayload, queryParams)

	var errorErr error
	if errorPayload != nil {
		errorErr = blrsvc.SendPullError(errorPayload, queryParams)
	}

	var errorMap map[string]error = make(map[string]error)
	if dataErr != nil {
		errorMap["data"] = dataErr
	}
	if errorErr != nil {
		errorMap["error"] = errorErr
	}
	if len(errorMap) > 0 {
		return fmt.Errorf("failed to send data and error: %v", errorMap)
	}
	return nil
}

var bluelockRelayService = di.NewThreadSafeSingleton(func() *BluelockRelayService {
	cfg := config.AcquireConfig()
	relayBaseURL := cfg.Common.RelayBaseURL
	apiKey := cfg.Secrets.DDApiKey
	orgCode := cfg.Common.OrgCode
	activeIntegrationService := cfg.ActiveService
	return NewBluelockRelayService(relayBaseURL, orgCode, activeIntegrationService, apiKey)
})

func AcquireBluelockRelayService() *BluelockRelayService {
	return bluelockRelayService.Acquire()
}
