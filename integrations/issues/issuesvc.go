package issues

import "github.com/bluelock-go/integrations"

type IssueIntegrationService interface {
	integrations.IntegrationService
	// IssuePull fetches issues from the issue tracking system
	IssuePull() error
	// IssueActivityPull fetches issue activity such as comments, transitions, etc.
	IssueActivityPull() error
	// IssueMetricsPull fetches issue metrics such as time tracking, velocity, etc.
	IssueMetricsPull() error
}

// ensureIssueIntegrationServiceImplementation enforces at compile time that all required issue integration services implement the IssueIntegrationService interface.
func ensureIssueIntegrationServiceImplementation() {
	// Add issue services here as they are implemented
	// var _ IssueIntegrationService = (*jira.JiraSvc)(nil)
}

// init performs package-level initialization and enforces compile-time checks for IssueIntegrationService implementations.
func init() {
	ensureIssueIntegrationServiceImplementation()
}
