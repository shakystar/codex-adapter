# codex-adapter

[문서 색인](docs/README.md)에서 설계·검증 기록과 문서 공백을 확인할 수 있습니다.

로컬에 설치된 Codex app-server를 호출하는 Go 어댑터입니다. Python, JavaScript, Rust, shell 등에서 JSON CLI로 사용할 수 있습니다.

어댑터는 stdio로 로컬 Codex 프로세스를 시작합니다. ChatGPT 인증과 모델 추론에는 원격 서비스 연결이 필요합니다.

<a id="build-and-check"></a>

## 빌드와 점검

Go 1.26 이상과 인증된 Codex CLI가 필요합니다. 프로토콜은 0.166.0과 0.169.1에서 확인했습니다.

```sh
go test -race ./...
go vet ./...
go build -o bin/codex-adapter ./cmd/codex-adapter
./bin/codex-adapter --doctor
```

`--doctor`는 모델 턴을 시작하거나 토큰을 강제로 갱신하지 않고 `account/read`를 호출합니다. 계정 조회 성공만으로 모델 호출 가능 여부를 판단할 수 없으며, 라우팅 시간 초과만으로 rate limiting을 확정할 수 없습니다. `workspace_routing_timeout`의 재현 조건은 [Docker 주소 비교 및 임시 설정](docs/validation.md#docker-workspace-routing-check-2026-11-20)에 있습니다.

<a id="execute-from-any-language"></a>

## CLI 호출

stdin으로 JSON 요청 하나를 전달합니다. CLI는 stdout에 결과 이벤트를, stderr에 크기를 제한하고 민감 정보를 제거한 상태 이벤트를 기록합니다. 오류 시 종료 코드는 0이 아닙니다. 같은 `sessionKey`와 `--state` 디렉터리를 사용하면 명령을 다시 시작한 뒤에도 대화를 재개할 수 있습니다.

```sh
./bin/codex-adapter --state ./adapter-state <<'JSON'
{"id":"task-1","sessionKey":"reviewer-project-a","workspace":"/absolute/project/path","instructions":"Review the supplied text.","tools":["advisory"],"input":"Explain the tradeoffs of this design: ..."}
JSON
```

`--model`로 해당 턴의 모델을 지정할 수 있습니다. 결과 이벤트에는 토큰 사용량이 포함됩니다. [JSON 계약과 Python 예제](docs/protocol.md)를 참고할 수 있습니다. Go에서는 `github.com/shakystar/codex-adapter`를 가져와 `New(Config)`와 `Execute(ctx, Request)`를 호출합니다. 저수준 stdio RPC는 `/rpc`에서 제공합니다.

<a id="behavior-and-limits"></a>

## 동작과 제한

- 인스턴스마다 프로세스 하나를 유지하며 한 번에 한 호출자가 실행을 소유합니다.
- 호출자가 지정한 세션 식별자, thread 매핑, 중단된 턴의 실행 의도를 저장합니다. 실행 여부가 불확실한 작업은 자동 재전송하지 않습니다.
- 읽기 전용 `advisory`와 공개 웹 조사 프로필을 지원합니다. 프로필은 운영 모드이며 OS 격리를 제공하지 않습니다. 지원하지 않는 도구 알림은 수신 후 거부합니다.
- Heartbeat는 로컬 RPC의 응답 여부를 나타냅니다. 모델 진행률은 측정하지 않습니다. 모델 턴에는 기본 경과 시간 제한이 없으며 제어 요청에는 제한이 있습니다.
- 실행 전 계정 검사에서 관리형 인증을 확인합니다. 외부 토큰 갱신 콜백, 대화형 승인, 임의의 동적 도구는 지원하지 않습니다.
- RPC 레코드, 알림, stderr, 입력, 출력의 크기를 제한합니다. 사용량은 성공한 턴만 집계하며 일부 값이 누락될 수 있습니다.
- 별도 인스턴스나 프로세스가 같은 상태 디렉터리를 동시에 사용하면 안 됩니다. 스케줄링, 잠금, 작업 큐, 재시도와 복구 정책은 호출 애플리케이션이 관리합니다.

이 저장소는 role-runtime에 의존하지 않습니다. 런타임 연결 코드는 role-runtime의 `internal/codexadapter` 패키지에 있습니다. [검증 기록](docs/validation.md)에서 자동 검사와 6개 역할 실모델 복구 실행의 결과 및 한계를 확인할 수 있습니다.

<a id="failure-handling"></a>

## 실패 처리

`Failure` 메타데이터는 작업 실패 시 모델 요청이 어디까지 전달됐는지 나타냅니다. 시작 중 취소를 지원하려면 `NewContext`를 사용합니다. 계정 사전 검사에서 일시적 오류로 분류된 경우에만 제한적으로 재시도합니다. 턴 제출은 자동 재시도하지 않습니다. `--state DIR --inspect-session KEY`로 모델 호출 없이 로컬 실행 기록을 조회합니다. [실패 계약과 구성요소별 책임](docs/protocol.md#failure-contract-v1-and-responsibility-2026-11-19)에 상세 내용을 정리했습니다.