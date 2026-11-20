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

## Docker workspace routing check (2026-11-20)

On macOS and Linux/amd64 Docker, Codex CLI 0.170.0 returned `workspace routing discovery timed out` on two direct `account/read` calls per environment. Each call used `refreshToken=false` and took about 15 seconds. The Linux CLI recognized the ChatGPT login with both a read-only mounted auth file and a writable temporary copy. Both setups exhibited the timeout. The same error was already recorded on November 18, before the November 20 Python caller integration.

We compared the two IPv4 addresses returned by DNS for `chatgpt.com`. HTTPS probes used the original hostname and normal certificate verification. For the account check, two disposable containers used the same image, Codex binary, Go doctor and read-only credentials. Only the per-container address mapping differed.

| DNS address observed | HTTPS from macOS and Docker | Go adapter doctor in Docker |
| --- | --- | --- |
| `104.18.32.47` | 2/2 HTTP 200 in each environment, all under 0.5 s | 2/2 returned `routingAvailable=true` |
| `172.64.155.209` | 2/2 TCP connect timeouts in each environment, with a 5 s limit | 2/2 returned `workspace_routing_timeout` |

For a one-run check, resolve `chatgpt.com` again, then test each current IPv4 address with the following command. Set `IP` to one address at a time.

```sh
curl --ipv4 --resolve "chatgpt.com:443:${IP}" --connect-timeout 5 --max-time 8 \
  --silent --show-error --output /dev/null \
  --write-out 'HTTP %{http_code}, TCP %{time_connect}s, total %{time_total}s\n' https://chatgpt.com/
```

Add `--add-host="chatgpt.com:${IP}"` to a single `docker run --rm` using a responding address from the current lookup. Run `codex-adapter --doctor` in that container. The mapping expires with the container. This comparison identified address-dependent TCP reachability from the test network. The failing network hop and the cause of earlier incidents remain unknown.

**Live turn with the temporary mapping.** At 04:35 UTC a fresh Linux/amd64 container used `--add-host=chatgpt.com:104.18.32.47` and Codex CLI 0.170.0. The Go doctor returned `routingAvailable=true`. In the same container, agent-eval then executed one planned M3 solo maintenance-notice task through role-orchestrator and role-runtime with the Codex backend. The writer completed in 5.41 seconds of recorded run time. The runtime receipt reported `succeeded`, its output matched the saved artifact byte-for-byte, and the request matched the recorded attempt ID. The 127-byte output described the checkout outage, continued browsing and retry time. The artifact-presence rule passed (one judgment, zero scorer errors). Output SHA-256 was `54fb607f980029e0d93eae0187e34bd2be470a4abfed5258cff44af58b183ff0`. This follow-up covered one solo turn.
