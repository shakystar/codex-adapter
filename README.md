# codex-adapter

Reusable adapter for a locally installed Codex app-server. The implementation is Go; the JSON command-line contract can be called from Python, JavaScript, Rust, shell, or other languages.

The adapter starts Codex locally over stdio. ChatGPT authentication and model inference still use remote services. This is not an offline model or a hosted Codex service.

## Build and check

Requires Go 1.26+ and an authenticated Codex CLI (protocol checked against 0.166.0 and 0.169.1).

```sh
go test -race ./...
go vet ./...
go build -o bin/codex-adapter ./cmd/codex-adapter
./bin/codex-adapter --doctor
```

`--doctor` calls `account/read` without a model turn or forced token refresh. Successful account lookup does not prove model readiness. A routing timeout is not proof of rate limiting.

## Execute from any language

Provide one JSON request on stdin. The command writes a result event to stdout and bounded, sanitized status events to stderr. It exits nonzero on error. Use a stable `sessionKey` and the same `--state` directory to resume a logical conversation, including after restarting the command.

```sh
./bin/codex-adapter --state ./adapter-state <<'JSON'
{"id":"task-1","sessionKey":"reviewer-project-a","workspace":"/absolute/project/path","instructions":"Review the supplied text.","tools":["advisory"],"input":"Explain the tradeoffs of this design: ..."}
JSON
```

See [the JSON contract](docs/protocol.md). Go callers can import `github.com/shakystar/codex-adapter` and use `New(Config)` / `Execute(ctx, Request)`. Raw stdio RPC is available at `/rpc`.

## Behavior and limits

- Persistent process per adapter instance; one execution owner at a time.
- Caller-supplied session identity, durable thread mapping, explicit interrupted-turn intent. No automatic replay of an uncertain task.
- Read-only advisory or public web research profiles. These are operating modes, not an OS security boundary. Unsupported tool notifications are rejected after observation.
- Heartbeat reports local RPC responsiveness, not model progress. Model turns have no default elapsed-time limit; control requests are bounded.
- Account preflight checks managed authentication before execution. External token refresh callbacks, interactive approvals, and arbitrary dynamic tools are not supported.
- Bounded RPC records, notifications, stderr, input, and output. Usage reports cover successful turns only and can be incomplete.
- Separate instances/processes must not share the same state directory concurrently. The application owns scheduling, locking, work queues, retries, and recovery policy.

This repository does not depend on role-runtime. Its runtime integration remains in that project's `internal/codexadapter` package. [Validation notes](docs/validation.md) distinguish automated checks from the successful six-role live recovery run and its limits.
