# Dependencies and provenance

The Go module uses the standard library only. Initial RPC and session handling were extracted from shakystar/role-runtime, based on its 5f06913-era implementation and subsequent local work. No Codex executable or credentials are included.

Runtime integration: separately installed OpenAI Codex CLI 0.166.0 / 0.169.1. Its upstream project is https://github.com/openai/codex (Apache-2.0 source repository); installed binary redistribution terms were not reviewed because this project does not redistribute it. The wire schema is experimental and can change. Replace the executable using `--codex` / rpc.Config.Command; alternate backends require a separate implementation.
