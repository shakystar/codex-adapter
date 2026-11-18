# Validation status

Extracted from role-runtime's Codex integration. Tests use fake app-server subprocesses and cover bounded transport, cancellation, heartbeat, role/session identity, and process restart restoration. role-runtime retains its integration suite for scheduler and durable-store behavior.

Live backend recovery is still being investigated. Prior account/read calls intermittently returned `workspace routing discovery timed out`; both original and isolated local environments also had successful account lookups. No evidence yet proves a local working-directory defect or rate limiting. A passing unit test or doctor call is not proof of a successful live model turn or full research pipeline.

CLI versions observed: 0.166.0 and 0.169.1. The Codex executable is not redistributed. Managed ChatGPT authentication is supported; externally managed token refresh is not. Consult current OpenAI authentication terms before commercial or hosted use.
