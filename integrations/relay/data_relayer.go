package relay

import "net/url"

type DataRelayer interface {
	SendCollectedData(payload interface{}, queryParams url.Values) error
	SendPullError(payload interface{}, queryParams url.Values) error

	// If data payload is nil, return an error with message "data payload is nil". If error payload is nil, do not send pull error.
	SendDataAndError(dataPayload interface{}, errorPayload interface{}, queryParams url.Values) error
}

var _ DataRelayer = (*BluelockRelayService)(nil)
