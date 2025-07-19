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

// ensureCICDIntegrationServiceImplementation is a placeholder for compile-time checks to ensure that CI/CD service implementations satisfy the CICDIntegrationService interface.
func ensureCICDIntegrationServiceImplementation() {
	// Add CI/CD services here as they are implemented
	// var _ CICDIntegrationService = (*jenkins.JenkinsSvc)(nil)
}

// init ensures that CICDIntegrationService interface implementation checks are performed during package initialization.
func init() {
	ensureCICDIntegrationServiceImplementation()
}
