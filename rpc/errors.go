package codexrpc

import (
	"encoding/json"
	"strings"
)

// ClassifyMessage returns a closed vocabulary; raw remote text is never retained.
func ClassifyMessage(message string) string {
	s := strings.ToLower(message)
	switch {
	case strings.Contains(s, "workspace routing discovery timed out"):
		return "workspace_routing_timeout"
	case strings.Contains(s, "workspace routing discovery unauthorized"):
		return "workspace_routing_unauthorized"
	case strings.Contains(s, "too many requests"):
		return "rate_limited"
	case strings.Contains(s, "usage limit"):
		return "usage_limit"
	default:
		return "unknown"
	}
}
func (e *RPCError) UnmarshalJSON(data []byte) error {
	var wire struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	e.Code = wire.Code
	e.Reason = ClassifyMessage(wire.Message)
	return nil
}
