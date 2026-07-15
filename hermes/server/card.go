package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	cardColorInfo    = "#2196F3"
	cardColorSuccess = "#4CAF50"
	cardColorWarning = "#FF9800"
	cardColorError   = "#F44336"
)

func commandCardResponse(title, text, color, footer string) *model.CommandResponse {
	resp := &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         strings.TrimSpace(text),
	}
	attachCardProps(resp, title, text, color, footer)
	return resp
}

func postWithCard(channelID, rootID, message, title, color, footer string) *model.Post {
	post := &model.Post{
		ChannelId: channelID,
		RootId:    rootID,
		Message:   strings.TrimSpace(message),
	}
	attachPostCardProps(post, title, message, color, footer)
	return post
}

func attachCardProps(resp *model.CommandResponse, title, text, color, footer string) {
	resp.Props = map[string]any{
		"attachments": []map[string]any{mattermostAttachment(title, text, color, footer)},
	}
}

func attachPostCardProps(post *model.Post, title, text, color, footer string) {
	post.Props = map[string]any{
		"attachments": []map[string]any{mattermostAttachment(title, text, color, footer)},
	}
}

func mattermostAttachment(title, text, color, footer string) map[string]any {
	attachment := map[string]any{
		"text": strings.TrimSpace(text),
	}
	if strings.TrimSpace(title) != "" {
		attachment["pretext"] = strings.TrimSpace(title)
	}
	if strings.TrimSpace(color) != "" {
		attachment["color"] = strings.TrimSpace(color)
	}
	if strings.TrimSpace(footer) != "" {
		attachment["footer"] = strings.TrimSpace(footer)
	}
	return attachment
}

func acceptedCardText(commandText string) string {
	commandText = strings.TrimSpace(commandText)
	if commandText == "" {
		return "Hermes 명령을 접수했습니다. 완료되면 이 채널/스레드에 비공개 응답으로 전달합니다."
	}
	return fmt.Sprintf("Hermes 명령을 접수했습니다.\n\n```text\n/hermes %s\n```", commandText)
}

func resultCardTitle(commandText string) string {
	commandText = strings.TrimSpace(commandText)
	if commandText == "" {
		return "✅ Hermes 응답"
	}
	return "✅ Hermes 응답 · /hermes " + commandText
}
