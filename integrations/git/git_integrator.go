package git

import "github.com/bluelock-go/integrations"

type GitIntegrator interface {
	integrations.Integrator
	// RepoPull fetches the repositories from VCS
	RepoPull() error
	// GitActivityPull fetches the activity from VCS such as commits, pull requests, reviews, etc.
	GitActivityPull() error
}

// PriorityScheduledGitIntegrator is used when code breakdown data is sent via a separate scheduled job,
// not during the regular job run.
type PriorityScheduledGitIntegrator interface {
	GitIntegrator
	// GitCodeBreakdownPull fetches the diffs from VCS such as lines of code, files changed, etc.
	GitCodeBreakdownPull() error
}
