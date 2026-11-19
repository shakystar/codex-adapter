package codexadapter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	codexrpc "github.com/shakystar/codex-adapter/rpc"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCodexAdapterProcess(t *testing.T) {
	if os.Getenv("ROLE_RUNTIME_ADAPTER_HELPER") != "1" {
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	send := func(v any) {
		if err := encoder.Encode(v); err != nil {
			os.Exit(2)
		}
	}
	var activeID, mode string
	beats := 0
	complete := func() {
		send(map[string]any{"method": "turn/completed", "params": map[string]any{"threadId": activeID, "turn": map[string]any{"id": "turn-1", "status": "completed", "items": []any{map[string]any{"type": "agentMessage", "phase": "final_answer", "text": "delayed success"}}}}})
	}
	for scanner.Scan() {
		var req struct {
			ID     json.RawMessage            `json:"id"`
			Method string                     `json:"method"`
			Params map[string]json.RawMessage `json:"params"`
		}
		if json.Unmarshal(scanner.Bytes(), &req) != nil {
			os.Exit(3)
		}
		if len(req.ID) == 0 {
			continue
		}
		if os.Getenv("ROLE_RUNTIME_STALL_CONTROL") == req.Method {
			continue
		}
		if path := os.Getenv("CODEX_ADAPTER_CALL_LOG"); path != "" {
			f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if e != nil {
				os.Exit(10)
			}
			fmt.Fprintln(f, req.Method)
			f.Close()
		}
		failAccount := os.Getenv("CODEX_ADAPTER_AUTH_ERROR") == "1"
		if marker := os.Getenv("CODEX_ADAPTER_FAIL_ONCE"); marker != "" && req.Method == "account/read" {
			f, e := os.OpenFile(marker, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if e == nil {
				f.Close()
				failAccount = true
			}
		}
		if req.Method == "account/read" && failAccount {
			send(map[string]any{"id": req.ID, "error": map[string]any{"code": -32603, "message": func() string {
				if os.Getenv("CODEX_ADAPTER_UNAUTHORIZED") == "1" {
					return "workspace routing discovery unauthorized private-data"
				}
				return "workspace routing discovery timed out private-data"
			}()}})
			continue
		}
		var result any = map[string]any{}
		switch req.Method {
		case "account/read":
			result = map[string]any{"account": map[string]any{"type": "chatgpt"}, "requiresOpenaiAuth": true}
		case "thread/loaded/list":
			beats++
			if mode == "probe-dies" {
				os.Exit(0)
			}
			if mode == "probe-stalls" {
				continue
			}
			if mode == "complete-during-probe" {
				complete()
				time.Sleep(60 * time.Millisecond)
				mode = ""
			}
			if mode == "heartbeat-delayed" && beats == 3 {
				send(map[string]any{"method": "thread/status/changed", "params": map[string]any{"threadId": activeID, "status": map[string]any{"type": "active", "activeFlags": []string{"waitingOnUserInput"}}}})
				complete()
			}
			if mode == "progress-delayed" {
				for i := 0; i < 25; i++ {
					send(map[string]any{"method": "item/reasoning/textDelta", "params": map[string]any{"threadId": activeID, "turnId": "turn-1", "delta": "private-reasoning-must-not-appear-in-status"}})
					time.Sleep(time.Millisecond)
				}
				complete()
			}
			result = map[string]any{"data": []string{activeID}}
		case "thread/start", "thread/resume":
			var cwd, instructions, id string
			json.Unmarshal(req.Params["cwd"], &cwd)
			json.Unmarshal(req.Params["baseInstructions"], &instructions)
			json.Unmarshal(req.Params["threadId"], &id)
			var config map[string]any
			json.Unmarshal(req.Params["config"], &config)
			wantSearch := "disabled"
			if strings.Contains(instructions, "Web research task:") {
				wantSearch = "live"
			}
			if config["web_search"] != wantSearch || config["features.shell_tool"] != false {
				os.Exit(9)
			}
			if id == "" {
				id = "thread-" + digest(instructions)[:12]
			}
			result = map[string]any{"thread": map[string]any{"id": id, "cwd": cwd}}
		case "turn/start":
			var id string
			var input []struct {
				Text string `json:"text"`
			}
			json.Unmarshal(req.Params["threadId"], &id)
			json.Unmarshal(req.Params["input"], &input)
			activeID = id
			mode = input[0].Text
			beats = 0
			if input[0].Text == "require-schema" {
				var schema map[string]any
				if json.Unmarshal(req.Params["outputSchema"], &schema) != nil || schema["type"] != "object" || schema["additionalProperties"] != false {
					send(map[string]any{"id": req.ID, "error": map[string]any{"code": -32602}})
					continue
				}
			}
			if input[0].Text == "lost-ack" {
				os.Exit(0)
			}
			send(map[string]any{"id": req.ID, "result": map[string]any{"turn": map[string]any{"id": "turn-1", "status": "inProgress"}}})
			if input[0].Text == "hang" {
				continue
			}
			if strings.HasPrefix(mode, "probe-") || strings.HasSuffix(mode, "-delayed") || mode == "complete-during-probe" {
				continue
			}
			notify := func(method string, p any) { send(map[string]any{"method": method, "params": p}) }
			var webFinalItems []any
			if mode == "web-tool" || mode == "web-limit" || mode == "web-no-id" {
				webItem := func(itemID, action string) map[string]any {
					return map[string]any{"id": itemID, "type": "webSearch", "query": "private-query-secret", "action": map[string]any{"type": action, "url": "https://private-url-secret.example"}, "results": []any{map[string]any{"private": "private-result-secret"}}}
				}
				for _, pair := range [][2]string{{"other-thread", "turn-1"}, {id, "other-turn"}} {
					notify("item/completed", map[string]any{"threadId": pair[0], "turnId": pair[1], "item": webItem("irrelevant", "search")})
				}
				if mode == "web-limit" {
					for i := 0; i < 257; i++ {
						webFinalItems = append(webFinalItems, webItem(fmt.Sprint(i), "search"))
					}
				} else if mode == "web-no-id" {
					webFinalItems = []any{webItem("", "search")}
				}
				w := webItem("search-1", "search")
				notify("item/started", map[string]any{"threadId": id, "turnId": "turn-1", "item": w})
				for i := 0; i < 2; i++ {
					notify("item/completed", map[string]any{"threadId": id, "turnId": "turn-1", "item": w})
				}
				if mode == "web-tool" {
					webFinalItems = []any{w, webItem("open-1", "openPage"), webItem("find-1", "findInPage"), webItem("other-1", "private-action-secret")}
				}
			}
			if mode == "backend-errors" {
				for i, code := range []any{"serverOverloaded", map[string]any{"responseStreamDisconnected": map[string]any{"httpStatusCode": 503, "private": "private-error-secret"}}, "private-error-secret"} {
					notify("error", map[string]any{"threadId": id, "turnId": "turn-1", "willRetry": i != 2, "error": map[string]any{"codexErrorInfo": code, "message": "private-error-secret", "additionalDetails": "private-error-secret", "misalignment": map[string]any{"detailedExplanation": "private-error-secret"}}})
				}
			}
			if mode == "string-status" {
				notify("mcpServer/startupStatus/updated", map[string]any{"threadId": id, "status": "ready", "delta": map[string]any{"other": "shape"}, "turn": "unrelated"})
			}
			if input[0].Text == "tool" {
				notify("item/started", map[string]any{"threadId": id, "turnId": "turn-1", "item": map[string]any{"type": "commandExecution"}})
				continue
			}
			notify("thread/tokenUsage/updated", map[string]any{"threadId": id, "turnId": "turn-1", "tokenUsage": map[string]any{"last": map[string]any{"inputTokens": 10, "outputTokens": 3, "cachedInputTokens": 2}}})
			status := "completed"
			if input[0].Text == "fail" {
				status = "failed"
			}
			notify("turn/completed", map[string]any{"threadId": id, "turn": map[string]any{"id": "turn-1", "status": status, "items": append(webFinalItems, map[string]any{"type": "agentMessage", "phase": "final_answer", "text": id + ":" + input[0].Text})}})
			continue
		case "thread/unsubscribe":
			if mode == "cleanup-lost" {
				os.Exit(0)
			}
			result = map[string]any{"status": "unsubscribed"}
		}
		send(map[string]any{"id": req.ID, "result": result})
	}
	os.Exit(0)
}

func TestIndependentSessionRestore(t *testing.T) {
	state := t.TempDir()
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{StateDir: state, RPC: codexrpc.Config{Command: os.Args[0], Args: []string{"-test.run=^TestCodexAdapterProcess$"}, Env: append(os.Environ(), "ROLE_RUNTIME_ADAPTER_HELPER=1")}}
	a, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := Request{ID: "one", SessionKey: "alpha", Workspace: workspace, Instructions: "alpha", Tools: []string{"advisory"}, Input: "hello"}
	first, err := a.Execute(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	req.SessionKey = "beta"
	req.Instructions = "beta"
	req.ID = "two"
	second, err := a.Execute(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if first.Session == second.Session || first.PID != second.PID {
		t.Fatal("role isolation/process reuse failed")
	}
	a.Close()
	a, err = New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	req.SessionKey = "alpha"
	req.Instructions = "alpha"
	req.ID = "three"
	third, err := a.Execute(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if third.Session != first.Session {
		t.Fatal("session failed to restore")
	}
}
func TestRejectTraversal(t *testing.T) {
	for _, key := range []string{"../escape", "/absolute", "", ".."} {
		r := Request{ID: "id", SessionKey: key, Workspace: "/tmp", Input: "hello"}
		if r.Validate() == nil {
			t.Fatalf("accepted key %q", key)
		}
	}
}

func TestAccountFailureBeforeExecution(t *testing.T) {
	cfg := Config{StateDir: t.TempDir(), RPC: codexrpc.Config{Command: os.Args[0], Args: []string{"-test.run=^TestCodexAdapterProcess$"}, Env: append(os.Environ(), "ROLE_RUNTIME_ADAPTER_HELPER=1", "CODEX_ADAPTER_AUTH_ERROR=1")}}
	a, err := New(cfg)
	if a != nil {
		a.Close()
		t.Fatal("preflight failure created executable adapter")
	}
	if err == nil || !strings.Contains(err.Error(), "workspace_routing_timeout") || strings.Contains(err.Error(), "private-data") {
		t.Fatalf("wrong safe diagnosis: %v", err)
	}
}
