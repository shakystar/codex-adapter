# JSON command contract (v0)

Implementation language does not constrain callers. This is a single-request CLI contract, not a network server or an MCP server. The CLI manages one local Codex process per invocation; the Go API can reuse a process across many requests.

## Invocation

`codex-adapter --codex /path/to/codex --state /path/to/state [--model MODEL]`

When supplied, `--model` sets the model on thread start and resume. Omitting the flag uses the configured Codex model.

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

Success stdout contains one JSON object.

```json
{"event":"result","result":{"text":"...","session":"...","pid":123},"usage":{"scope":"successful-turn-last-notification","inputTokens":10,"outputTokens":3,"cachedInputTokens":2,"known":true,"completedTurns":1,"measuredTurns":1}}
```

The CLI reports usage for the successful turn in this invocation, including a resumed session. Check `known` and `measuredTurns` before using token counts. Missing usage can produce zero counts with `known:false`.

Progress stderr: `{"event":"status","status":{...}}`.
Failure stderr: `{"event":"error","error":"sanitized error"}` and exit code 1. Application output belongs in result.text and is not considered sanitized telemetry.

`--doctor` requires no stdin or state directory. It returns an `account` event containing only account type, authentication requirement, and routing availability. It does not expose account IDs, email, tokens, or backend addresses. RPC error reasons use a closed vocabulary; unknown remote messages remain `unknown`.

SIGINT/SIGTERM cancels execution. Failed or canceled model turns retain a durable intent; subsequent execution with that session fails closed. Do not erase this marker and automatically replay a task that may have performed external effects. Investigate the original run and apply an explicit recovery policy or create a new logical session.

## Failure contract v1 and responsibility (2026-11-19)

`New`/`NewContext` and `Execute` return an error discoverable with `FailureOf(err)`
(or `errors.As`). JSON CLI error events retain `event`/`error` and add `failure`:

```json
{"event":"error","error":"codex account_read: workspace_routing_timeout (not_started)","failure":{"version":1,"stage":"account_read","code":"workspace_routing_timeout","execution":"not_started","retryable":true,"attempts":3}}
```

`execution` describes model dispatch evidence, not task success:

| Value | Evidence / action |
| --- | --- |
| not_started | This invocation has not dispatched turn/start. Retry eligibility still depends on code and caller policy. |
| unknown | turn/start may have reached the server, or an interrupted intent exists. Never automatically resend. |
| started | A turn ID was acknowledged. Later failure does not imply absence of effects. |
| completed | A completed turn was observed. Cleanup/persistence can still fail; do not rerun to fix cleanup. |

Error text and serialized failure omit remote free text and local underlying cause.
`Unwrap` preserves the cause for local diagnostics and `errors.Is` cancellation checks.
Stages distinguish initialization, account lookup, configuration, thread loading,
intent persistence, turn submission/observation, cleanup and recovery. Unknown errors
are `operation_failed`; authentication, cancellation and deadlines are distinguished.
The existing doctor diagnostic is still a single read-only request, not a model-readiness guarantee.

Startup has one total `StartupTimeout` budget (default 30 seconds), including bounded
backoff and process initialization. `StartupAttempts` counts all tries (default 3,
maximum 5; set 1 to disable retries). `RetryBackoff` defaults to 250 ms, doubles up to
5 seconds, and is cancelable with `NewContext`. Only classified routing timeout and
rate limiting in preflight are retried; authorization, unknown errors and model calls
are not. Every unsuccessful initialized client is closed before another attempt.
A `startup_retry` status carries sanitized failure metadata. Exhaustion returns the
last structured error; `retryable` indicates eligibility, not automatic retry permission.
Callers should avoid multiplying retry loops across layers.

Turn submission never retries. A durable intent precedes dispatch; acknowledged turn
IDs and execution evidence are persisted. `InspectSession(state, sessionKey)` or
`--state DIR --inspect-session KEY` reads that evidence without starting Codex, changing
state or exposing prompts. Legacy intents without execution evidence are unknown.
Inspection is **local evidence**, not a current remote turn query or a recovery command.
A crash before the acknowledgment is persisted may leave less precise evidence.
Intent remains until successful cleanup; no API silently clears it or resends a turn.

Adapter owns protocol failure classification and bounded startup recovery. Runtime owns
receipts, session quarantine and deduplication. Orchestrator owns scheduling/waiting/
operator decisions. No external routing-service repair or exactly-once guarantee is claimed.

## Python caller

The standard library can call the CLI and retain the terminal error event alongside any status messages.

```python
import json
import subprocess


def run_turn(binary, state, request, model):
    command = [binary, "--state", str(state), "--model", model]
    proc = subprocess.run(command, input=json.dumps(request), capture_output=True,
                          text=True, encoding="utf-8")
    if proc.returncode:
        for line in reversed(proc.stderr.splitlines()):
            try:
                event = json.loads(line)
            except ValueError:
                continue
            if isinstance(event, dict) and event.get("event") == "error":
                return event
        raise RuntimeError("Adapter stopped without a terminal error event")
    return json.loads(proc.stdout)
```

Pass the request fields in the table above. Read `event["result"]["text"]` and `event["usage"]` on success. For an error, inspect `event["failure"]["execution"]` when available. A missing terminal event leaves dispatch uncertain.
