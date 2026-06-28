# Hermes 프로젝트 운영 정책

생성/점검 시각: 2026-06-28 19:32:36 KST
대상: 이 프로젝트에서 Hermes가 작업할 때 우선 적용할 공통 정책

## 1. 기본 응답/작업 원칙
- 한국어로 응답한다.
- 첫 문장은 결론 또는 현재 상태를 먼저 말한다.
- 설명은 실무형으로 간결하게 하되, 검증 결과와 근거는 명확히 남긴다.
- 불확실한 내용은 추정으로 단정하지 않고 `현재 데이터 기준`, `추가 확인 필요`처럼 범위를 표시한다.
- 사용자가 문제를 지적하면 방어적으로 설명하지 말고, 즉시 확인·수정·재검증한다.

## 2. 실행/검증 원칙
- “하겠습니다”라고 말한 작업은 같은 턴에서 실제 도구 호출로 수행한다.
- 파일 내용, 시스템 상태, 날짜/시간, git 상태, 빌드/테스트 결과는 기억이나 추정으로 답하지 않고 도구로 확인한다.
- 코드를 수정하면 가능한 범위에서 실제 빌드/테스트/설정 검증을 수행한다.
- 실패한 도구 결과나 막힌 경로는 꾸며내지 않고 그대로 보고하고 대안을 시도한다.
- 위험하거나 운영 영향이 있는 재기동/배포/삭제는 범위와 영향을 확인한 뒤 진행한다.

## 3. 파일/명령 사용 규칙
- 파일 읽기: `read_file` 우선.
- 파일 검색: `search_files` 우선.
- 파일 수정: `patch` 또는 `write_file` 우선.
- shell은 빌드, 테스트, git, 프로세스, Docker, 네트워크 확인처럼 shell이 필요한 경우에 사용한다.
- 여러 독립 확인이 필요하면 병렬 도구 호출을 우선한다.

## 4. Obsidian 기록/정리 정책
- Vault: `/Users/iseungchan/Library/Mobile Documents/iCloud~md~obsidian/Documents/Note`
- Obsidian 정리는 PARA 역할 경계를 지킨다.
  - 프로젝트 진행, 구현/설계/task/status, 작업 산출물 → `1. Project/<project>/`
  - 개발 일정, 실행 이력, 운영 작업 로그 → `2. Area/EventNote/`
  - 재사용 가능한 개념, 패턴, 기술 설명, 조사/합성 지식 → `3. Resource/<domain>/`
  - 용어, 약어, 프로토콜/객체명, vocabulary/definition → `4. Archive/Dictionary/`
- raw source는 원문성을 보존하고, LLM 생성 요약/합성 지식과 섞지 않는다.
- 개발 작업이 프로젝트 구조, API, 배포 토폴로지, 검증 계약, 메뉴/대시보드 구성을 바꾸면 EventNote와 별개로 `1. Project/<project>/`의 안정적인 구조 노트도 갱신한다.
- reusable knowledge가 작업 중 생기면 `1. Project`에만 묻어두지 말고 `3. Resource`로 승격하고, 필요하면 Project 노트에서 링크한다.
- 새 용어·약어·프로토콜명·도메인 객체명이 반복 사용될 때는 기존 Dictionary를 먼저 검색한 뒤 `4. Archive/Dictionary/`에 생성/갱신한다.
- EventNote frontmatter는 `type: single`, `date`, quoted `startTime`/`endTime`, `timezone: Asia/Seoul`, `allDay: false`, `project`, `status`, `tags`를 사용한다.
- 같은 주제라도 시간이 다른 후속 작업은 기존 노트에 합치지 않고 별도 EventNote로 분리한다.
- Obsidian 파일 작성/수정 후에는 대표 파일을 읽기 검증한다.
- Vault 변경 후 LiveSync는 가능한 경우 `command id=obsidian-livesync:livesync-scan-files`, `command id=obsidian-livesync:livesync-runbatch`, `command id=obsidian-livesync:livesync-replicate` 순서로 실행하고 결과를 확인한다.

## 5. NAS/운영 환경 정책
- NAS 파일 확인은 가능한 경우 로컬 마운트 `/Users/iseungchan/NAS_Shares/docker/...`를 우선 사용한다.
- SSH/Docker 직접 조작은 마운트로 불가능한 Docker/DB/런타임 확인이 필요할 때만 사용한다.
- NAS `docker-compose.yml`, nginx, runtime 파일 변경은 운영 영향이 있으므로 수정 전후 diff/문법 검증을 수행한다.
- compose 변경은 `docker compose config --quiet`로 검증한다.
- nginx 변경은 컨테이너 nginx `nginx -t`로 검증한다.
- 실행 중 컨테이너에 반영하려면 reload/recreate 필요 여부를 별도로 보고한다.

## 6. 코드 변경 정책
- 공통 모듈이나 타 프로젝트 재사용 코드 수정 시 영향 범위를 먼저 검토한다.
- 기존 동작을 바꾸는 경우 회귀 테스트 또는 최소 재현 검증을 추가/수정한다.
- 로컬에 Java 등 런타임이 없으면 Docker 기반 검증처럼 재현 가능한 대체 경로를 사용한다.
- 빌드 산출물 생성 여부와 명령 결과를 실제 출력 기준으로 보고한다.

## 7. Git/배포 정책
- git 상태, diff, 브랜치, 커밋 이력은 반드시 도구로 확인한다.
- 커밋/푸시/배포는 사용자의 의도와 운영 영향을 확인한 뒤 진행한다.
- 임시 파일, 백업, 런타임 산출물이 섞이지 않도록 diff 범위를 점검한다.

## 8. 민감정보/기록 정책
- 비밀번호, 토큰, secret, 개인식별 민감정보는 문서나 메모에 평문 기록하지 않는다.
- 장기 메모에는 1주일 내 stale될 작업 진행상황, PR/issue 번호, 커밋 해시, 임시 TODO를 저장하지 않는다.
- 반복 절차는 memory보다 skill 또는 프로젝트 정책 문서에 남긴다.

## 9. 프로젝트별 특화 정책 — mattermost-plugin
- 이 저장소는 Mattermost `/hermes` slash command plugin과 Hermes Agent API 연결을 관리한다.
- 변경 시 `hermes/` 하위 plugin 코드를 우선 확인하고, 가능하면 `make -C hermes test`와 `make -C hermes bundle`로 검증한다.
- Mattermost slash command 응답은 사용자에게 노출되는 운영 인터페이스이므로 에러 메시지, timeout, API endpoint, 인증/토큰 처리를 명확히 검증한다.
- 배포 산출물(`dist/`, bundle 결과)은 의도된 release artifact인지 확인하고, 임시 빌드 산출물을 무조건 커밋하지 않는다.
- Hermes API 호출 경로 변경 시 Mattermost DM/채널 호출 규칙과 `@hermes` 알림 정책에 영향을 주는지 확인한다.

## 10. Hermes 정책 갱신 원칙
- Hermes 공통 운영 정책을 추가/수정할 때는 이 파일의 공통 섹션만 바꾸지 말고, 반드시 위의 프로젝트별 특화 정책과 충돌/보완 필요성을 함께 점검한다.
- 다른 프로젝트에 정책을 전파할 때도 단순 복사 대신 각 repo의 언어, 런타임, 배포 방식, 민감정보 위치, 기존 AGENTS/CLAUDE/GEMINI 지침을 반영해 `.agent/HERMES_POLICIES.md`를 조정한다.
- 기존 `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`가 있으면 `.agent/HERMES_POLICIES.md`가 대체하지 않고 공존하며, 더 구체적인 repo-local 지침을 우선한다.
