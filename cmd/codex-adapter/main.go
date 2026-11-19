// codex-adapter exposes the same execution contract to callers in any language.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	adapter "github.com/shakystar/codex-adapter"
	rpc "github.com/shakystar/codex-adapter/rpc"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		json.NewEncoder(os.Stderr).Encode(map[string]any{"event": "error", "error": err.Error(), "failure": adapter.FailureOf(err)})
		os.Exit(1)
	}
}
func run() error {
	binary := flag.String("codex", "codex", "installed Codex executable")
	state := flag.String("state", "", "persistent adapter state directory (required for execution)")
	inspect := flag.String("inspect-session", "", "read local session evidence without starting Codex")
	doctor := flag.Bool("doctor", false, "check account without a model turn")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *inspect != "" {
		evidence, err := adapter.InspectSession(*state, *inspect)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "inspection", "inspection": evidence})
	}
	cfg := rpc.Config{Command: *binary}
	if *doctor {
		cfg.Args = []string{"app-server", "--stdio", "-c", "web_search=\"disabled\"", "-c", "features.apps=false", "-c", "features.plugins=false"}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		c, err := rpc.Start(ctx, cfg)
		if err != nil {
			return err
		}
		defer c.Close()
		status, err := adapter.CheckAccount(ctx, c)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "account", "status": status})
	}
	data, err := io.ReadAll(io.LimitReader(os.Stdin, 256<<10+1))
	if err != nil {
		return err
	}
	if len(data) > 256<<10 {
		return fmt.Errorf("request exceeds 256KiB")
	}
	var req adapter.Request
	if err = json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("invalid request JSON")
	}
	if err = req.Validate(); err != nil {
		return err
	}
	a, err := adapter.NewContext(ctx, adapter.Config{RPC: cfg, StateDir: *state, OnStatus: func(s adapter.Status) {
		json.NewEncoder(os.Stderr).Encode(map[string]any{"event": "status", "status": s})
	}})
	if err != nil {
		return err
	}
	defer a.Close()
	result, err := a.Execute(ctx, req)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "result", "result": result})
}
