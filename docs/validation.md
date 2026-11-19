# Validation status

Extracted from role-runtime's Codex integration. Tests use fake app-server subprocesses and cover bounded transport, cancellation, heartbeat, role/session identity, and process restart restoration. role-runtime retains its integration suite for scheduler and durable-store behavior.

Live recovery was validated in one complete six-role run on 2026-11-19; sustained availability and the earlier intermittent routing timeout cause remain unproven. Prior account/read calls intermittently returned `workspace routing discovery timed out`; both original and isolated local environments also had successful account lookups. No evidence proves a local working-directory defect or rate limiting. A passing unit test or doctor call is not proof of a successful live model turn or full research pipeline.

CLI versions observed: 0.166.0 and 0.169.1. The Codex executable is not redistributed. Managed ChatGPT authentication is supported; externally managed token refresh is not. Consult current OpenAI authentication terms before commercial or hosted use.

## Live recovery probe, 2026-11-18

The published adapter commit 2e2a0da completed a single advisory model turn via the language-independent JSON CLI using Codex 0.169.1. The expected literal `RECOVERY_OK` matched the actual final output; turn execution and thread unload took about 3.76 seconds. No tools were used. This proves one successful live execution, not sustained availability or a repaired upstream routing service. A user-canceled run completed four roles; later runs exposed a runtime wall-clock lease issue, now fixed with monotonic deadlines and regression tests. A subsequent run stopped at account preflight with workspace_routing_timeout before any model turn. These failures remain in the runtime experiment history; a later full run succeeded as recorded below.

A controlled comparison of inherited versus newly created process groups succeeded in all three account lookups (0.50/0.30/0.45 seconds). Matching Python initialization parameters and string RPC IDs to the Go client also succeeded, followed by a successful Go doctor call in the same execution environment. These results do not support a deterministic process-group or JSON-RPC formatting defect; earlier intermittent failures remain documented.

## Complete pipeline verification, 2026-11-19

role-runtime commit 12bd987, using this adapter's implementation commit 2e2a0da and installed Codex 0.169.1, completed six specialist roles on two workers in 1357.13 seconds. Five real SMS waits resumed successfully; all six durable task receipts completed. The two research turns overlapped. Four research roles recorded 15 web-tool completions in total, and advisory roles recorded none. There were no observed backend errors.

Independent artifact checks verified six document hashes, six SMS artifact references, and all five predecessor references in the final decision input. This validates transport and evidence handoff, not the truth of model research or the commercial viability of its decision. See [the runtime report](https://github.com/shakystar/role-runtime/blob/main/experiments/event-search/target-review-team.md).

The host had AC idle sleep configured to one minute. This run used process-scoped `caffeinate -i`; its idle-sleep assertion was confirmed active and released automatically on exit. Global power settings were unchanged. A separately approved account request returned HTTP 200, and the official app-server doctor succeeded before the run. Sleep is a plausible contributor to earlier interruptions, but this does not establish the cause of every routing timeout.

## Failure responsibility hardening — 2026-11-19

Synthetic real-stdio subprocess tests verify transient preflight recovery, three-attempt
exhaustion, nonretryable authorization failure, cancellation during backoff, sanitized
JSON failure fields, lost turn acknowledgment, acknowledged failed turns, successful
completion evidence and no replay after reopening. Call logs assert zero turn/start
requests during startup recovery and exactly one for lost-ack/failed-turn scenarios.
Existing session, heartbeat, tool and RPC tests remain part of the race suite.
These are injected failures, not evidence that the upstream routing service was repaired.
The local inspection API is not remote reconciliation; uncertain intents remain blocked.
