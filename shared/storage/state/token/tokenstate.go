package token

import "time"

// TokenState represents the state of a single token
type TokenState struct {
	LastUsageAt              time.Time   `json:"lastUsageAt"`
	ExhaustedAt              time.Time   `json:"exhaustedAt"`
	Status                   TokenStatus `json:"status"`
	StatusChangedAt          time.Time   `json:"statusChangedAt"`
	SuccessfulUsageCount     int         `json:"successfulUsageCount"`
	PreRateLimitSuccessCount int         `json:"preRateLimitSuccessCount"`
}

func (ts *TokenState) IsExhausted() bool {
	return ts.Status == TokenExhausted
}
func (ts *TokenState) IsUnauthorized() bool {
	return ts.Status == TokenUnauthorized
}
func (ts *TokenState) IsActive() bool {
	return ts.Status == TokenActive
}
func (ts *TokenState) UpdateTokenStatus(status TokenStatus, statusChangedAt time.Time) {
	ts.Status = status
	ts.StatusChangedAt = statusChangedAt
}

func (ts *TokenState) SetTokenAsExhausted(exhaustionTime time.Time) {

	ts.UpdateTokenStatus(TokenExhausted, exhaustionTime)
	ts.ExhaustedAt = exhaustionTime
}

func (ts *TokenState) UpdateTokenUsage(tokenUsageTime time.Time) {
	ts.LastUsageAt = tokenUsageTime
	ts.SuccessfulUsageCount++
}

func (ts *TokenState) ResetUsageMetrics(resumeTime time.Time) {
	ts.UpdateTokenStatus(TokenActive, resumeTime)
	ts.PreRateLimitSuccessCount = ts.SuccessfulUsageCount
	ts.SuccessfulUsageCount = 0
}
