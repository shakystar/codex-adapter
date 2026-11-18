# JSON command contract (v0)

Implementation language does not constrain callers. This is a single-request CLI contract, not a network server or an MCP server. The CLI manages one local Codex process per invocation; the Go API can reuse a process across many requests.

## Invocation

`codex-adapter --codex /path/to/codex --state /path/to/state`

Read a single JSON object from stdin (maximum 256 KiB), then close stdin. No secret credentials belong in the request. Authentication is handled by the installed Codex client.

| Field | Meaning |
| --- | --- |
| id | Nonempty task identity, at most 256 bytes |
| sessionKey | Stable conversation identity; ASCII letters/digits followed by letters/digits/underscore/dot/hyphen, at most 128 characters |
| role | Optional diagnostic label |
| workspace | Absolute local working directory |
| instructions | Caller-resolved instructions, at most 64 KiB |
| tools | Exactly one profile: `advisory` or `web-research` |
| input | Nonempty input, at most 64 KiB |

Instructions and workspace must match the saved mapping when resuming a session. Changing them requires a new session key. The adapter adds profile instructions; keep application roles and work queues outside this API.

Success stdout: `{"event":"result","result":{"text":"...","session":"...","pid":123}}`.
Progress stderr: `{"event":"status","status":{...}}`.
Failure stderr: `{"event":"error","error":"sanitized error"}` and exit code 1. Application output belongs in result.text and is not considered sanitized telemetry.

`--doctor` requires no stdin or state directory. It returns an `account` event containing only account type, authentication requirement, and routing availability. It does not expose account IDs, email, tokens, or backend addresses. RPC error reasons use a closed vocabulary; unknown remote messages remain `unknown`.

SIGINT/SIGTERM cancels execution. Failed or canceled model turns retain a durable intent; subsequent execution with that session fails closed. Do not erase this marker and automatically replay a task that may have performed external effects. Investigate the original run and apply an explicit recovery policy or create a new logical session.
