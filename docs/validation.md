# Validation status

Extracted from role-runtime's Codex integration. Tests use fake app-server subprocesses and cover bounded transport, cancellation, heartbeat, role/session identity, and process restart restoration. role-runtime retains its integration suite for scheduler and durable-store behavior.

Live backend recovery is still being investigated. Prior account/read calls intermittently returned `workspace routing discovery timed out`; both original and isolated local environments also had successful account lookups. No evidence yet proves a local working-directory defect or rate limiting. A passing unit test or doctor call is not proof of a successful live model turn or full research pipeline.

CLI versions observed: 0.166.0 and 0.169.1. The Codex executable is not redistributed. Managed ChatGPT authentication is supported; externally managed token refresh is not. Consult current OpenAI authentication terms before commercial or hosted use.

## Live recovery probe, 2026-11-18

The published adapter commit 2e2a0da completed a single advisory model turn via the language-independent JSON CLI using Codex 0.169.1. The expected literal `RECOVERY_OK` matched the actual final output; turn execution and thread unload took about 3.76 seconds. No tools were used. This proves one successful live execution, not sustained availability or a repaired upstream routing service. Six-role validation in role-runtime remains incomplete: a user-canceled run completed four roles; later runs exposed a runtime wall-clock lease issue, now fixed with monotonic deadlines and regression tests. The latest run stopped at account preflight with workspace_routing_timeout before any model turn. No pipeline is currently running.

A controlled comparison of inherited versus newly created process groups succeeded in all three account lookups (0.50/0.30/0.45 seconds). Matching Python initialization parameters and string RPC IDs to the Go client also succeeded, followed by a successful Go doctor call in the same execution environment. These results do not support a deterministic process-group or JSON-RPC formatting defect; earlier intermittent failures remain documented.
