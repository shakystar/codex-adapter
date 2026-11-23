## 개발 기준과 검증

[공통 개발 정책](https://github.com/shakystar/agent-collaboration-lab/blob/c541f5ba978d06bc0e5ce7a1c1df0dff1166149f/docs/development-policy.md)을 적용한다. 검증 명령은 레포 루트에서 실행한다. go.mod에 지정된 Go 환경에서 `go vet ./...` 및 `go test -race ./...`
