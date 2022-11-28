package pipeline

// Describes a polaris-scheduler Status.
type StatusCode int

const (
	// Success means that the plugin has executed correctly and deemed the pod to be schedulable.
	// A nil Status is also considered as a Success.
	Success StatusCode = iota

	// Designates an internal plugin error, such as unexpected input, etc.
	// This should NOT be used when a pod is deemed to be unschedulable.
	InternalError

	// Unschedulable means that the plugin cannot find a node (within the plugin's scope) to place the pod.
	// The Reasons array should be set to the reason for the unschedulability.
	Unschedulable
)

// String representations of the StatusCodes - these must use the same order as the const definitions above.
var statusCodeStrings = []string{
	"Success",
	"InternalError",
	"Unschedulable",
}

// Reports the result of a pod's journey through the Polaris scheduling pipeline.
// A nil Status is also considered a success.
type Status interface {

	// Gets the StatusCode.
	Code() StatusCode

	// Gets the StatusCode as a string.
	CodeAsString() string

	// Gets the error that occurred, if any, otherwise returns nil.
	Error() error

	// Gets array of reasons for the current status.
	// This may also be nil.
	Reasons() []string

	// Gets the plugin that has caused scheduling to fail.
	// This is set by the framework and is nil, if all plugins returned success.
	FailedPlugin() Plugin

	// Gets the stage of the scheduling pipeline that caused scheduling to fail.
	// This is set by the framework and is an empty string, if all plugins returned success.
	FailedStage() string

	// Sets the the plugin that has caused the scheduling pipeline to fail.
	// This should be done by the scheduling pipeline only.
	SetFailedPlugin(plugin Plugin, stage string)

	// Gets the reasons for the current status as a single string.
	Message() string
}

// Returns true if the status represents a Success, otherwise false.
// A nil status also represents a Success.
func IsSuccessStatus(status Status) bool {
	return status == nil || status.Code() == Success
}

// Returns the string version of the specified Status' statusCode.
// This also works if status is nil (which also represents a Success status).
func StatusCodeAsString(status Status) string {
	if status != nil {
		return status.CodeAsString()
	} else {
		return statusCodeStrings[Success]
	}
}
