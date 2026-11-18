package codexadapter

import (
	"context"
	"errors"
	codexrpc "github.com/shakystar/codex-adapter/rpc"
)

// AccountStatus excludes tokens, identifiers, email, and backend URLs.
type AccountStatus struct {
	Type             string `json:"type,omitempty"`
	RequiresAuth     bool   `json:"requiresAuth"`
	RoutingAvailable bool   `json:"routingAvailable"`
}

// CheckAccount does not start a model turn or force token refresh.
func CheckAccount(ctx context.Context, c *codexrpc.Client) (AccountStatus, error) {
	var reply struct {
		Account *struct {
			Type string `json:"type"`
		} `json:"account"`
		RequiresAuth *bool     `json:"requiresOpenaiAuth"`
		Routing      *struct{} `json:"workspaceRouting"`
	}
	if err := c.Call(ctx, "account/read", map[string]bool{"refreshToken": false}, &reply); err != nil {
		return AccountStatus{}, err
	}
	if reply.RequiresAuth == nil {
		return AccountStatus{}, errors.New("account response missing requiresOpenaiAuth")
	}
	status := AccountStatus{RequiresAuth: *reply.RequiresAuth, RoutingAvailable: reply.Routing != nil}
	if reply.Account != nil {
		switch reply.Account.Type {
		case "chatgpt", "apiKey", "amazonBedrock", "gcp":
			status.Type = reply.Account.Type
		default:
			status.Type = "unknown"
		}
	}
	if status.RequiresAuth && reply.Account == nil {
		return status, errors.New("Codex login required")
	}
	return status, nil
}
