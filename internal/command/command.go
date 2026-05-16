package command

import (
	"context"
	"strings"
)

const (
	ResponseEphemeral = "ephemeral"

	UsageText = "사용법: /hermes compress"
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

	name := strings.Fields(strings.TrimSpace(req.Text))
	if len(name) == 0 || name[0] != "compress" {
		return ephemeral(UsageText)
	}

	result, err := h.gateway.Execute(ctx, GatewayCommand{
		Name:      "compress",
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

func ephemeral(text string) SlashResponse {
	return SlashResponse{ResponseType: ResponseEphemeral, Text: text}
}
