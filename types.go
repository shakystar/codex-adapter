package codexadapter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"regexp"
)

// Request is independent of any scheduler, role registry, or storage engine.
// SessionKey identifies one logical conversation and must remain stable on resume.
type Request struct {
	ID           string   `json:"id"`
	Role         string   `json:"role,omitempty"`
	SessionKey   string   `json:"sessionKey"`
	Workspace    string   `json:"workspace"`
	Instructions string   `json:"instructions"`
	Tools        []string `json:"tools"`
	Input        string   `json:"input"`
}
type TurnUsage struct {
	InputTokens       int64 `json:"input_tokens"`
	OutputTokens      int64 `json:"output_tokens"`
	CachedInputTokens int64 `json:"cached_input_tokens"`
}

type Result struct {
	Text      string     `json:"text,omitempty"`
	Session   string     `json:"session,omitempty"`
	PID       int        `json:"pid,omitempty"`
	TurnID    string     `json:"turn_id,omitempty"`
	TurnUsage *TurnUsage `json:"turn_usage,omitempty"`
}

var sessionKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)

func (r Request) Validate() error {
	if !sessionKeyPattern.MatchString(r.SessionKey) || r.ID == "" || len(r.ID) > 256 {
		return errors.New("invalid request identity")
	}
	if !filepath.IsAbs(r.Workspace) {
		return errors.New("workspace must be an absolute local path")
	}
	if len(r.Input) == 0 || len(r.Input) > 64<<10 || len(r.Instructions) > 64<<10 {
		return errors.New("input or instructions exceed bounds")
	}
	return nil
}
func digest(parts ...string) string {
	data, _ := json.Marshal(parts)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
