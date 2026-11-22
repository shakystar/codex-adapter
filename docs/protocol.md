<a id="json-command-contract-v0"></a>

# JSON 명령 계약 (v0)

호출 언어와 관계없이 CLI는 요청을 한 번씩 처리합니다. 네트워크 서버나 MCP 서버 기능은 제공하지 않습니다. CLI는 호출마다 로컬 Codex 프로세스 하나를 관리하고, Go API에서는 같은 프로세스를 여러 요청에 재사용할 수 있습니다.

<a id="invocation"></a>

## 호출 방법

`codex-adapter --codex /path/to/codex --state /path/to/state [--model MODEL]`

`--model`을 지정하면 thread 시작 및 재개 시 해당 모델을 사용합니다. 생략하면 Codex에 설정된 모델을 사용합니다.

stdin에 최대 256 KiB의 JSON 객체 하나를 전달한 뒤 stdin을 닫습니다. 인증은 설치된 Codex 클라이언트가 처리합니다. 요청에 인증 정보를 넣지 않습니다.

| 필드 | 의미 |
| --- | --- |
| id | 비어 있지 않은 작업 식별자. 최대 256바이트 |
| sessionKey | 고정 대화 식별자. 첫 글자는 ASCII 영문자 또는 숫자이며, 이후에는 영문자·숫자·밑줄·마침표·하이픈을 사용할 수 있습니다. 최대 128자 |
| role | 선택적 진단 레이블 |
| workspace | 로컬 작업 디렉터리의 절대 경로 |
| instructions | 호출자가 확정한 지침. 최대 64 KiB |
| tools | `advisory` 또는 `web-research` 중 하나 |
| input | 비어 있지 않은 입력. 최대 64 KiB |

세션을 재개할 때 지침과 작업 디렉터리가 저장된 매핑과 일치해야 합니다. 변경 시 새 세션 키를 사용합니다. 어댑터는 프로필 지침을 추가합니다. 애플리케이션의 역할과 작업 큐는 호출자가 관리합니다.

성공 시 stdout에 JSON 객체 하나를 출력합니다.

```json
{"event":"result","result":{"text":"...","session":"...","pid":123},"usage":{"scope":"successful-turn-last-notification","inputTokens":10,"outputTokens":3,"cachedInputTokens":2,"known":true,"completedTurns":1,"measuredTurns":1}}
```

CLI는 재개된 세션을 포함해 이번 호출에서 성공한 턴의 사용량을 보고합니다. 토큰 수를 사용하기 전에 `known`과 `measuredTurns`를 확인합니다. 사용량이 누락되면 `known:false`와 함께 토큰 수가 0으로 표시될 수 있습니다.

진행 중에는 stderr에 `{"event":"status","status":{...}}`를 기록합니다. 실패 시 stderr에 `{"event":"error","error":"sanitized error"}`를 기록하고 종료 코드 1을 반환합니다. 애플리케이션 출력은 result.text에 들어가며 민감 정보를 제거한 진단 데이터로 취급할 수 없습니다.

`--doctor`에는 stdin이나 상태 디렉터리가 필요하지 않습니다. `account` 이벤트에는 계정 유형, 인증 필요 여부, 라우팅 가능 여부만 포함됩니다. 계정 ID·이메일·토큰·백엔드 주소는 노출하지 않습니다. RPC 오류 사유는 미리 정의된 값으로 제한하며, 분류할 수 없는 원격 메시지는 `unknown`으로 기록합니다.

SIGINT/SIGTERM은 실행을 취소합니다. 실패하거나 취소한 모델 턴은 실행 의도(intent)를 저장하며, 같은 세션의 후속 실행을 차단합니다. 이미 외부에 영향을 주었을 수 있으므로 이 기록을 지워 작업을 자동 재전송하면 안 됩니다. 원래 실행을 조사한 뒤 명시적으로 복구하거나 새 논리 세션을 생성합니다.

<a id="failure-contract-v1-and-responsibility-2026-11-19"></a>

## 실패 계약 v1과 구성요소별 책임 (2026-11-19)

`New`/`NewContext`와 `Execute`가 반환한 오류는 `FailureOf(err)` 또는 `errors.As`로 확인할 수 있습니다. JSON CLI의 오류 이벤트는 기존 `event`와 `error`를 유지하고 `failure`를 추가합니다.

```json
{"event":"error","error":"codex account_read: workspace_routing_timeout (not_started)","failure":{"version":1,"stage":"account_read","code":"workspace_routing_timeout","execution":"not_started","retryable":true,"attempts":3}}
```

`execution`은 모델 턴 전달 상태를 나타냅니다.

| 값 | 확인된 상태와 처리 방법 |
| --- | --- |
| not_started | 이번 호출에서는 turn/start를 전송하지 않았습니다. 재시도 가능 여부는 오류 코드와 호출자 정책에 따라 결정합니다. |
| unknown | turn/start가 서버에 도달했을 수 있거나 중단된 실행 의도가 남아 있습니다. 자동 재전송하지 않습니다. |
| started | 턴 ID를 확인했습니다. 이후 실패했어도 이미 외부에 영향을 주었을 수 있습니다. |
| completed | 턴 완료를 관찰했습니다. 정리나 상태 저장에 실패했더라도 정리를 위해 턴을 다시 실행하지 않습니다. |

오류 문구와 직렬화된 실패 정보에서는 원격 자유 형식 텍스트와 로컬 내부 원인을 제외합니다. `Unwrap`은 로컬 진단과 `errors.Is`를 통한 취소 확인에 필요한 원인을 보존합니다. 단계별로 초기화, 계정 조회, 구성, thread 로드, 실행 의도 저장, 턴 제출·관찰, 정리, 복구를 구분합니다. 분류할 수 없는 오류는 `operation_failed`로 기록하고 인증·취소·기한 초과는 구분합니다. doctor는 읽기 전용 요청 한 건을 수행하며 모델 실행 가능 여부까지 보장하지 않습니다.

시작 과정은 backoff와 프로세스 초기화를 포함해 하나의 `StartupTimeout`을 공유합니다(기본 30초). `StartupAttempts`는 최초 시도를 포함합니다(기본 3회, 최대 5회, 1로 설정하면 재시도 없음). `RetryBackoff`는 기본 250 ms에서 두 배씩 늘어나 최대 5초이며, `NewContext`로 취소할 수 있습니다. 사전 검사에서 라우팅 시간 초과 또는 rate limiting으로 분류된 경우만 재시도합니다. 권한 오류, 알 수 없는 오류, 모델 호출은 재시도하지 않습니다. 초기화한 클라이언트에서 실패하면 다음 시도 전에 해당 클라이언트를 닫습니다. `startup_retry` 상태에는 민감 정보를 제거한 실패 메타데이터를 담습니다. 횟수를 소진하면 마지막 구조화 오류를 반환합니다. `retryable`은 재시도 가능성을 나타내며 자동 재시도를 허가하지 않습니다. 호출 계층마다 재시도 루프가 중첩되지 않도록 합니다.

턴 제출은 재시도하지 않습니다. 전송 전에 실행 의도를 저장하고 확인된 턴 ID와 실행 증거를 기록합니다. `InspectSession(state, sessionKey)` 또는 `--state DIR --inspect-session KEY`로 Codex 시작, 상태 변경, 프롬프트 노출 없이 해당 기록을 확인할 수 있습니다. 실행 증거가 없는 이전 형식의 intent는 unknown으로 처리합니다. 조회 결과는 **로컬에 저장된 증거**이며 현재 원격 상태를 조회하거나 복구하지 않습니다. 승인 응답을 저장하기 전에 프로세스가 종료되면 증거의 정밀도가 낮아질 수 있습니다. 정리가 성공할 때까지 intent를 유지하며, 이를 암묵적으로 지우거나 턴을 재전송하는 API는 없습니다.

어댑터는 프로토콜 실패 분류와 제한된 시작 복구를 담당합니다. 런타임은 영수증, 세션 격리, 중복 제거를 담당합니다. 오케스트레이터는 스케줄링, 대기, 운영자 판단을 담당합니다. 외부 라우팅 서비스의 복구나 exactly-once 실행은 보장하지 않습니다.

<a id="python-caller"></a>

## Python에서 호출

Python 표준 라이브러리로 CLI를 호출하고 상태 메시지와 마지막 오류 이벤트를 함께 보존할 수 있습니다.

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

위 표의 필드를 요청에 담습니다. 성공하면 `event["result"]["text"]`와 `event["usage"]`를 읽고, 실패하면 `event["failure"]["execution"]`이 있는지 확인합니다. 마지막 이벤트가 누락되면 턴 전달 여부는 불확실합니다.