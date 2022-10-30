package podsubmission

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
)

// DTO for error responses.
type PodSubmissionApiError struct {
	Error *client.PolarisErrorDto `json:"error" yaml:"error"`
}
