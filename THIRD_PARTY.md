<a id="dependencies-and-provenance"></a>

# 의존성과 출처

Go 모듈은 표준 라이브러리만 사용합니다. 초기 RPC 및 세션 처리 코드는 shakystar/role-runtime의 5f06913 시점 구현과 이후 로컬 작업에서 분리했습니다. Codex 실행 파일과 인증 정보는 배포물에 포함하지 않습니다.

실행 시 별도 설치한 OpenAI Codex CLI 0.166.0 / 0.169.1과 연동합니다. [OpenAI Codex 소스 저장소](https://github.com/openai/codex)의 라이선스는 Apache-2.0입니다. 이 프로젝트는 설치된 바이너리를 재배포하지 않아 바이너리 재배포 조건은 검토하지 않았습니다. 통신 스키마는 실험 단계이며 변경될 수 있습니다. `--codex` 또는 `rpc.Config.Command`로 실행 파일을 지정합니다. 다른 백엔드를 사용하려면 별도 구현이 필요합니다.