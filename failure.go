package codexadapter

import (
	"context"
	"errors"
	"fmt"
	rpc "github.com/shakystar/codex-adapter/rpc"
)

const (
	NotStarted = "not_started"
	Unknown    = "unknown"
	Started    = "started"
	Completed  = "completed"
)

// Failure is a versioned, sanitized decision contract. Retryable is eligibility,
// not authorization to replay a task. Unwrap preserves the local cause.
type Failure struct {
	Version   int    `json:"version"`
	Stage     string `json:"stage"`
	Code      string `json:"code"`
	Execution string `json:"execution"`
	Retryable bool   `json:"retryable"`
	Attempts  int    `json:"attempts"`
	cause     error
}

func (f *Failure) Error() string {
	return fmt.Sprintf("codex %s: %s (%s)", f.Stage, f.Code, f.Execution)
}
func (f *Failure) Unwrap() error { return f.cause }

// ExecutionFailure lets schedulers consume metadata without importing this package.
func (f *Failure) ExecutionFailure() (string, string, string, bool) {
	return f.Stage, f.Code, f.Execution, f.Retryable
}
func FailureOf(err error) *Failure {
	var f *Failure
	if errors.As(err, &f) {
		return f
	}
	return nil
}
func failure(err error, stage, state string, attempts int) error {
	if err == nil {
		return nil
	}
	if FailureOf(err) != nil {
		return err
	}
	code := "operation_failed"
	var remote *rpc.RPCError
	switch {
	case errors.Is(err, context.Canceled):
		code = "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		code = "timeout"
	case errors.Is(err, ErrLoginRequired):
		code = "authentication_required"
	case errors.As(err, &remote):
		code = remote.Reason
	}
	retry := state == NotStarted && (code == "workspace_routing_timeout" || code == "rate_limited")
	return &Failure{Version: 1, Stage: stage, Code: code, Execution: state, Retryable: retry, Attempts: attempts, cause: err}
}

var ErrLoginRequired = errors.New("Codex login required")
