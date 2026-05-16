# Mattermost Hermes Slash Plugin

Mattermost `/hermes` slash command를 Hermes Gateway 명령으로 전달하는 플러그인입니다.

## 현재 MVP

- `/hermes compress` 라우팅
- Mattermost 사용자 allowlist 기반 권한 제어
- Hermes Gateway `POST /api/v1/mattermost/commands/compress` 호출
- 실패 시 민감정보를 응답에 노출하지 않는 안전한 ephemeral 응답

## 설정

Mattermost System Console의 플러그인 설정에서 다음 값을 지정합니다.

| Key | 설명 |
| --- | --- |
| `HermesGatewayURL` | Hermes Gateway base URL. 예: `http://127.0.0.1:8080` |
| `HermesGatewayToken` | Gateway 호출용 Bearer token |
| `AllowedUserIDs` | 쉼표로 구분한 Mattermost user ID allowlist. 비우면 전체 허용 |

## 개발

```bash
go test ./...
make bundle
```

번들 산출물은 `dist/dltmdcks98.hermes-slash-plugin.tar.gz`에 생성됩니다.
