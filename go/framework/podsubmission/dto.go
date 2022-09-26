package podsubmission

// DTO for error responses.
type PodSubmissionApiError struct {
	Error error `json:"error" yaml:"error"`
}
