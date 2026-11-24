package codexadapter

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewContextRejectsUnsupportedReasoningEffort(t *testing.T) {
	// Cancellation prevents a real startup attempt even if validation regresses.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, effort := range []string{"minimal", "xhigh", "HIGH", " low "} {
		t.Run(effort, func(t *testing.T) {
			a, err := NewContext(ctx, Config{
				ReasoningEffort: effort,
				OnProcessStart:  func(int) { t.Error("unexpected process start") },
			})
			cause := errors.Unwrap(err)
			if a != nil || cause == nil || !strings.Contains(cause.Error(), "invalid reasoning effort") {
				t.Fatalf("got adapter=%t, error=%v; want an effort validation error before startup", a != nil, err)
			}
		})
	}
}

func TestTurnStartParamsReasoningEffort(t *testing.T) {
	// This checks parameter construction only. Before rollout, verify each
	// model with low, medium, and high against the real app-server.
	for _, model := range []string{"gpt-6.2-sol", "gpt-6.2-astra"} {
		for _, effort := range []string{"", "low", "medium", "high"} {
			params := turnStartParams("thread", "input", Config{Model: model, ReasoningEffort: effort})
			got, present := params["effort"]
			if present != (effort != "") || (present && got != effort) {
				t.Errorf("model %s, configured effort %q: got effort=%v (present=%t)", model, effort, got, present)
			}
		}
	}
}
