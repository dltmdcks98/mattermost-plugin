# Mattermost Hermes Slash Plugin

Mattermost `/hermes` slash command를 Hermes Agent의 OpenAI-compatible API server로 전달하는 플러그인입니다.

## 현재 기능

- `/hermes <명령어 또는 질문>` 라우팅
- Mattermost 사용자 allowlist 기반 권한 제어
- Hermes Agent API `POST /v1/chat/completions` 호출
- `X-Hermes-Session-Key`로 Mattermost channel/thread 기준 세션 스코프 전달
- 접수/완료/오류 응답을 Mattermost attachment 카드 형태로 표시
- `/hermes help`, `/hermes status`, `/hermes model`, `/hermes new` 로컬 안내 응답
- `/hermes approve|deny|yolo`가 Gateway approval queue에 접근하지 못한다는 점을 명확히 안내
- 실패 시 민감정보를 응답에 노출하지 않는 안전한 ephemeral 응답

## Hermes Agent API 사전 설정

Hermes gateway에서 API server를 활성화해야 합니다.

```bash
hermes config set API_SERVER_ENABLED true
hermes config set API_SERVER_KEY '<secret>'
hermes config set API_SERVER_HOST 0.0.0.0
hermes config set API_SERVER_PORT 8642
hermes gateway stop && hermes gateway
```

Mattermost가 Docker 컨테이너에서 실행 중이면 플러그인 설정 URL은 보통 다음 값을 사용합니다.

```text
http://host.docker.internal:8642/v1
```

검증:

```bash
curl -s http://127.0.0.1:8642/health
curl -s -H 'Authorization: Bearer <secret>' http://127.0.0.1:8642/v1/models
```

## 플러그인 설정

Mattermost System Console의 플러그인 설정에서 다음 값을 지정합니다.

| Key | 설명 |
| --- | --- |
| `HermesAgentAPIURL` | Hermes Agent API URL. 예: `http://host.docker.internal:8642/v1` |
| `HermesAgentAPIKey` | Hermes `API_SERVER_KEY`와 동일한 Bearer token |
| `HermesAgentModel` | 모델명. 비우면 `hermes-agent` |
| `AllowedUserIDs` | 쉼표로 구분한 Mattermost user ID allowlist. 비우면 전체 허용 |

## 개발

```bash
go test ./...
make bundle
```

번들 산출물은 `dist/dltmdcks98.hermes-slash-plugin.tar.gz`에 생성됩니다.

> Mattermost 업로드용 tar.gz는 압축 해제 직후 루트에 `plugin.json`이 있어야 합니다. `make bundle` 산출물을 그대로 업로드하세요.
