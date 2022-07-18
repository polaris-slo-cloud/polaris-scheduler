package pipeline

import (
	"errors"
	"strings"
)

var (
	stImpl *statusImpl = nil

	_ Status = stImpl
)

// Default implementation of Status
type statusImpl struct {
	code         StatusCode
	err          error
	reasons      []string
	failedPlugin string
}

func NewStatus(code StatusCode, reasons ...string) Status {
	status := statusImpl{
		code:    code,
		reasons: reasons,
	}
	if code == InternalError {
		status.err = errors.New(status.Message())
	}
	return &status
}

func NewSuccessStatus() Status {
	return NewStatus(Success)
}

func NewInternalErrorStatus(err error) Status {
	status := statusImpl{
		code:    InternalError,
		err:     err,
		reasons: []string{err.Error()},
	}
	return &status
}

func (s *statusImpl) Code() StatusCode {
	return s.code
}

func (s *statusImpl) CodeAsString() string {
	return statusCodeStrings[s.code]
}

func (s *statusImpl) Error() error {
	return s.err
}

func (s *statusImpl) FailedPlugin() string {
	return s.failedPlugin
}

func (s *statusImpl) SetFailedPlugin(name string) {
	s.failedPlugin = name
}

func (s *statusImpl) Message() string {
	if s.reasons != nil {
		return strings.Join(s.reasons, ", ")
	}
	return ""
}

func (s *statusImpl) Reasons() []string {
	return s.reasons
}
