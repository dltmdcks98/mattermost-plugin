package command

import (
	"context"
	"strings"
)

const (
	ResponseEphemeral = "ephemeral"

	UsageText = "사용법: /hermes <명령어 또는 질문>"
)

type Config struct {
	AllowedUsers []string
}

type SlashRequest struct {
	UserID    string
	ChannelID string
	RootID    string
	Text      string
}

type SlashResponse struct {
	ResponseType string
	Text         string
}

type GatewayCommand struct {
	Name      string
	UserID    string
	ChannelID string
	RootID    string
}

type GatewayResult struct {
	Message string
}

type Gateway interface {
	Execute(ctx context.Context, cmd GatewayCommand) (GatewayResult, error)
}

type Handler struct {
	config  Config
	gateway Gateway
}

func NewHandler(config Config, gateway Gateway) *Handler {
	return &Handler{config: config, gateway: gateway}
}

func (h *Handler) Handle(ctx context.Context, req SlashRequest) SlashResponse {
	if !h.isAllowed(req.UserID) {
		return ephemeral("권한이 없습니다. Hermes slash 명령을 실행할 수 없습니다.")
	}

	name := strings.TrimSpace(req.Text)
	if name == "" {
		return ephemeral(UsageText)
	}
	if response, ok := gatewayOnlyCommandResponse(name); ok {
		return ephemeral(response)
	}

	result, err := h.gateway.Execute(ctx, GatewayCommand{
		Name:      name,
		UserID:    req.UserID,
		ChannelID: req.ChannelID,
		RootID:    req.RootID,
	})
	if err != nil {
		return ephemeral("Hermes 명령 실행에 실패했습니다. 잠시 후 다시 시도해주세요.")
	}
	if strings.TrimSpace(result.Message) == "" {
		return ephemeral("Hermes 명령이 완료되었습니다.")
	}
	return ephemeral(result.Message)
}

func (h *Handler) isAllowed(userID string) bool {
	if len(h.config.AllowedUsers) == 0 {
		return true
	}
	for _, allowed := range h.config.AllowedUsers {
		if allowed == userID {
			return true
		}
	}
	return false
}

func gatewayOnlyCommandResponse(name string) (string, bool) {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(name)))
	if len(fields) == 0 {
		return "", false
	}

	switch fields[0] {
	case "help", "commands":
		return strings.TrimSpace("🤖 **Hermes Mattermost Slash Plugin**\n\n" +
			"사용법: `/hermes <요청>`\n\n" +
			"자주 쓰는 예:\n" +
			"- `/hermes 오늘 작업 요약해줘`\n" +
			"- `/hermes FastAPI 상태 확인해줘`\n" +
			"- `/hermes 이 스레드 내용 기준으로 다음 액션 정리해줘`\n\n" +
			"주의:\n" +
			"- `/hermes`는 Hermes API 서버(`/v1/chat/completions`)로 일반 요청을 전달하는 slash bridge입니다.\n" +
			"- dangerous-command approval은 `/hermes approve`가 아니라 같은 Hermes 대화에서 `@hermes /approve`로 처리해야 합니다.\n" +
			"- `/model`, `/new` 같은 Gateway 세션 명령은 현재 slash bridge에서 직접 상태를 바꾸지 않습니다."), true
	case "status":
		return strings.TrimSpace("✅ **Hermes slash plugin is active**\n\n" +
			"이 플러그인은 Mattermost `/hermes` 명령을 Hermes Agent API server로 전달합니다.\n\n" +
			"확인 포인트:\n" +
			"- API URL/API Key는 Mattermost System Console의 플러그인 설정을 사용합니다.\n" +
			"- 완료 응답은 현재 채널/스레드에 비공개 카드로 표시됩니다.\n" +
			"- Gateway 전용 명령은 `@hermes /status`, `@hermes /approve`처럼 Gateway 대화 경로에서 실행하세요."), true
	case "model":
		return strings.TrimSpace("ℹ️ **모델 전환 안내**\n\n" +
			"참고 enhancer 프로젝트는 Thread별 모델 선택 카드를 제공합니다. 현재 Go slash plugin은 Hermes API bridge라 `/hermes model`만으로 Gateway 세션 모델을 직접 전환하지 않습니다.\n\n" +
			"현재 플러그인 기본 모델은 Mattermost System Console의 **HermesAgentModel** 설정값을 사용합니다. Thread별 모델 전환이 필요하면 enhancer식 callback/API 확장이 추가로 필요합니다."), true
	case "new", "reset":
		return strings.TrimSpace("ℹ️ **새 세션 안내**\n\n" +
			"현재 slash plugin은 Hermes API server의 session key를 채널/스레드 기준으로 고정해 연속성을 유지합니다.\n\n" +
			"새 대화가 필요하면 새 Mattermost Thread에서 `/hermes ...`를 실행하세요. Gateway 세션 리셋은 같은 Hermes 대화에서 `@hermes /new`를 사용하세요."), true
	case "approve", "deny", "yolo":
		return "이 명령은 Hermes Gateway의 대화 세션 안에서만 처리됩니다.\n\n" +
			"Mattermost `/hermes` 플러그인은 현재 OpenAI 호환 API(`/v1/chat/completions`)로 일반 요청만 전달하므로, 대기 중인 dangerous-command approval을 직접 조회하거나 해제할 수 없습니다.\n\n" +
			"승인하려면 승인 요청이 뜬 같은 Hermes 대화/스레드에서 Gateway 명령으로 처리해주세요:\n" +
			"- 1회 승인: `@hermes /approve`\n" +
			"- 전체 승인: `@hermes /approve all`\n" +
			"- 거절: `@hermes /deny`\n\n" +
			"참고: `/hermes approve`는 승인 명령이 아니라 이 플러그인의 API 경유 요청이라 approval 큐에 접근하지 못합니다.", true
	default:
		return "", false
	}
}

func ephemeral(text string) SlashResponse {
	return SlashResponse{ResponseType: ResponseEphemeral, Text: text}
}
