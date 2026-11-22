# 문서 색인

| 찾는 내용 | 원본 |
| --- | --- |
| 설치 조건, 빌드, 계정 점검, 호출 예시 | [시작 안내](../README.md) |
| CLI 요청·응답, 세션 재개, 실행 단계별 실패 처리, Python 호출 | [프로토콜](protocol.md) |
| 모의 서버 검사, 실모델 호출, 6역할 복구, Docker 라우팅 비교와 검증 한계 | [검증 기록](validation.md) |
| 구현 출처와 Codex CLI 의존성 | [의존성과 출처](../THIRD_PARTY.md) |

## 확인된 문서 공백

JSON CLI와 Python 예시는 있지만 Go 호출자가 초기화부터 실행·오류 확인·종료까지 따라갈 수 있는 예시는 없다. 현재 Go API의 동작 근거는 [어댑터 구현](../adapter.go), [오류 분류](../failure.go), [CLI 호출부](../cmd/codex-adapter/main.go)에 있다.

## 문서 관리

이 저장소는 어댑터 계약과 자체 검증 결과를 관리한다. 네 저장소의 통합 마일스톤 정본은 agent-eval의 `docs/milestones/vX.Y.Z/README.md`에 둔다. 작성 규칙은 [공통 문서 정책](https://github.com/shakystar/agent-collaboration-lab/blob/main/docs/documentation-policy.md)과 [마일스톤 양식](https://github.com/shakystar/agent-collaboration-lab/blob/main/templates/milestone.md)을 따른다. 두 링크는 비공개 저장소 접근 권한이 있어야 열 수 있다.
