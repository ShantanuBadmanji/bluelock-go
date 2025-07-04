package git

import "github.com/bluelock-go/integrations"

type GitIntegrationService interface {
	integrations.IntegrationService
	// RepoPull fetches the repositories from VCS
	RepoPull() error
	// GitActivityPull fetches the activity from VCS such as commits, pull requests, reviews, etc.
	GitActivityPull() error
}

// PriorityScheduledGitIntegrationService is used when code breakdown data is sent via a separate scheduled job,
// not during the regular job run.
type PriorityScheduledGitIntegrationService interface {
	GitIntegrationService
	// GitCodeBreakdownPull fetches the diffs from VCS such as lines of code, files changed, etc.
	GitCodeBreakdownPull() error
}
