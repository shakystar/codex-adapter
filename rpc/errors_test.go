package codexrpc

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestErrorRedaction(t *testing.T) {
	for _, tc := range []struct{ message, reason string }{
		{"workspace routing discovery timed out secret-token", "workspace_routing_timeout"},
		{"Too Many Requests secret-token", "rate_limited"},
		{"user@example.org secret-token", "unknown"},
	} {
		wire, _ := json.Marshal(map[string]any{"code": -32603, "message": tc.message, "data": "secret-token"})
		var err RPCError
		if e := json.Unmarshal(wire, &err); e != nil {
			t.Fatal(e)
		}
		encoded, _ := json.Marshal(err)
		if err.Reason != tc.reason || strings.Contains(err.Error(), "secret") || strings.Contains(string(encoded), "secret") {
			t.Fatal("unsafe classification")
		}
	}
}
