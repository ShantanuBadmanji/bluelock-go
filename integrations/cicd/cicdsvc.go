package cicd

import "github.com/bluelock-go/integrations"

type CICDIntegrationService interface {
	integrations.IntegrationService
	// BuildPull fetches build information from the CI/CD system
	BuildPull() error
	// PipelinePull fetches pipeline information
	PipelinePull() error
	// DeploymentPull fetches deployment information
	DeploymentPull() error
}

func ensureCICDIntegrationServiceImplementation() {
	// Add CI/CD services here as they are implemented
	// var _ CICDIntegrationService = (*jenkins.JenkinsSvc)(nil)
}

func init() {
	ensureCICDIntegrationServiceImplementation()
}
