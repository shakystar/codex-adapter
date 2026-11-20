package codexadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCLIContract(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "codex-adapter")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if output, err := exec.Command("go", "build", "-o", binary, "./cmd/codex-adapter").CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, output)
	}
	peer, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, model, input, stage, execution string }{
		{name: "default", input: "hello"},
		{name: "explicit", model: "gpt-6.2-sol", input: "hello"},
		{name: "bad-json", stage: "request_decode", execution: NotStarted},
		{name: "lost-ack", model: "gpt-6.2-sol", input: "lost-ack", stage: "turn_start", execution: Unknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			args := []string{"--codex", peer, "--state", t.TempDir()}
			if tc.model != "" {
				args = append(args, "--model", tc.model)
			}
			cmd := exec.CommandContext(ctx, binary, args...)
			cmd.Env = append(os.Environ(), "ROLE_RUNTIME_ADAPTER_HELPER=1", "CODEX_ADAPTER_TEST_MODEL="+tc.model)
			input := []byte("{")
			if tc.input != "" {
				input, err = json.Marshal(Request{ID: "test", SessionKey: "cli", Workspace: t.TempDir(), Instructions: "test", Tools: []string{"advisory"}, Input: tc.input})
				if err != nil {
					t.Fatal(err)
				}
			}
			cmd.Stdin = bytes.NewReader(input)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			stdout, err := cmd.Output()
			if tc.stage != "" {
				if err == nil || len(stdout) != 0 {
					t.Fatalf("error expected: stdout=%s err=%v", stdout, err)
				}
				lines := strings.Split(strings.TrimSpace(stderr.String()), "\n")
				var event struct {
					Event   string
					Failure *Failure
				}
				if err := json.Unmarshal([]byte(lines[len(lines)-1]), &event); err != nil {
					t.Fatal(err)
				}
				if event.Event != "error" || event.Failure == nil || event.Failure.Stage != tc.stage || event.Failure.Execution != tc.execution {
					t.Fatalf("failure: %+v", event)
				}
				return
			}
			if err != nil {
				t.Fatalf("execution: %v %s", err, stderr.String())
			}
			var result struct {
				Event  string
				Result Result
				Usage  Usage
			}
			if err := json.Unmarshal(stdout, &result); err != nil {
				t.Fatal(err)
			}
			if result.Event != "result" || !strings.HasSuffix(result.Result.Text, ":hello") || result.Result.Session == "" || result.Result.PID == 0 {
				t.Fatalf("result: %+v", result)
			}
			if !result.Usage.Known || result.Usage.CompletedTurns != 1 || result.Usage.InputTokens != 10 || result.Usage.OutputTokens != 3 || result.Usage.CachedInputTokens != 2 {
				t.Fatalf("usage: %+v", result.Usage)
			}
		})
	}
}
