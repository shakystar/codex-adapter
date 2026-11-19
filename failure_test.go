package codexadapter

import (
	"context"
	"encoding/json"
	"errors"
	rpc "github.com/shakystar/codex-adapter/rpc"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func failureConfig(t *testing.T, env ...string) Config {
	t.Helper()
	return Config{StateDir: t.TempDir(), StartupAttempts: 3, RetryBackoff: time.Millisecond, StartupTimeout: 3 * time.Second,
		RPC: rpc.Config{Command: os.Args[0], Args: []string{"-test.run=^TestCodexAdapterProcess$"}, Env: append(append(os.Environ(), "ROLE_RUNTIME_ADAPTER_HELPER=1"), env...)}}
}
func TestPreflightRetryBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name         string
		recover      bool
		unauthorized bool
		calls        int
	}{{"recover", true, false, 2}, {"exhausted", false, false, 3}, {"unauthorized", false, true, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			log := filepath.Join(t.TempDir(), "calls")
			env := []string{"CODEX_ADAPTER_CALL_LOG=" + log}
			if tc.recover {
				env = append(env, "CODEX_ADAPTER_FAIL_ONCE="+filepath.Join(t.TempDir(), "once"))
			} else {
				env = append(env, "CODEX_ADAPTER_AUTH_ERROR=1")
			}
			if tc.unauthorized {
				env = append(env, "CODEX_ADAPTER_UNAUTHORIZED=1")
			}
			cfg := failureConfig(t, env...)
			a, err := New(cfg)
			if tc.recover {
				if err != nil {
					t.Fatal(err)
				}
				a.Close()
			} else {
				f := FailureOf(err)
				if f == nil || f.Execution != NotStarted || f.Attempts != tc.calls || f.Retryable == tc.unauthorized {
					t.Fatalf("failure=%+v", f)
				}
				data, _ := json.Marshal(f)
				if strings.Contains(string(data), "private-data") {
					t.Fatal("raw remote data leaked")
				}
			}
			data, _ := os.ReadFile(log)
			if strings.Count(string(data), "account/read\n") != tc.calls || strings.Contains(string(data), "turn/start") {
				t.Fatalf("calls: %s", data)
			}
		})
	}
}
func TestStartupCancellationDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := failureConfig(t, "CODEX_ADAPTER_AUTH_ERROR=1")
	cfg.OnStatus = func(s Status) {
		if s.Phase == "startup_retry" {
			cancel()
		}
	}
	_, err := NewContext(ctx, cfg)
	f := FailureOf(err)
	if !errors.Is(err, context.Canceled) || f == nil || f.Execution != NotStarted || f.Retryable {
		t.Fatalf("%+v", err)
	}
}
func TestExecutionEvidenceAndNoReplay(t *testing.T) {
	for _, tc := range []struct{ input, state string }{{"lost-ack", Unknown}, {"fail", Started}, {"hello", Completed}} {
		t.Run(tc.input, func(t *testing.T) {
			log := filepath.Join(t.TempDir(), "calls")
			cfg := failureConfig(t, "CODEX_ADAPTER_CALL_LOG="+log)
			a, err := New(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			req := Request{ID: "t", SessionKey: "s", Workspace: t.TempDir(), Tools: []string{"advisory"}, Input: tc.input}
			_, err = a.Execute(context.Background(), req)
			if tc.state != Completed {
				f := FailureOf(err)
				if f == nil || f.Execution != tc.state || f.Retryable {
					t.Fatalf("failure=%+v", f)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			evidence, err := InspectSession(cfg.StateDir, "s")
			if err != nil {
				t.Fatal(err)
			}
			if evidence.Execution != tc.state || evidence.RecoveryRequired != (tc.state != Completed) {
				t.Fatalf("%+v", evidence)
			}
			if tc.state != Unknown && evidence.TurnID == "" {
				t.Fatal("turn id not persisted")
			}
			if tc.state != Completed {
				// A new process must not resend a previously dispatched turn.
				a.Close()
				a, err = New(cfg)
				if err != nil {
					t.Fatal(err)
				}
				defer a.Close()
				_, err = a.Execute(context.Background(), req)
				if FailureOf(err) == nil || FailureOf(err).Retryable {
					t.Fatal("unsafe recovery")
				}
			}
			data, _ := os.ReadFile(log)
			if strings.Count(string(data), "turn/start\n") != 1 {
				t.Fatalf("duplicate turn: %s", data)
			}
		})
	}
}
