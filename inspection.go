package codexadapter

import (
	"errors"
	"path/filepath"
)

// SessionInspection reports durable local evidence only. It does not contact the
// server, clear intent, or authorize replay. Unknown requires explicit recovery.
type SessionInspection struct {
	Version          int    `json:"version"`
	ThreadID         string `json:"threadId"`
	TaskID           string `json:"taskId,omitempty"`
	TurnID           string `json:"turnId,omitempty"`
	Execution        string `json:"execution"`
	RecoveryRequired bool   `json:"recoveryRequired"`
}

func InspectSession(state, key string) (SessionInspection, error) {
	if state == "" || !sessionKeyPattern.MatchString(key) {
		return SessionInspection{}, errors.New("invalid inspection identity")
	}
	m, err := readMapping(filepath.Join(state, "codex-sessions", key+".json"))
	if err != nil {
		return SessionInspection{}, err
	}
	if m.Key != key || m.ThreadID == "" {
		return SessionInspection{}, errors.New("invalid session mapping")
	}
	execution := m.Execution
	if execution != Started && execution != Completed {
		execution = Unknown
	}
	return SessionInspection{Version: 1, ThreadID: m.ThreadID, TaskID: m.ActiveTask, TurnID: m.TurnID, Execution: execution, RecoveryRequired: m.ActiveTask != ""}, nil
}
