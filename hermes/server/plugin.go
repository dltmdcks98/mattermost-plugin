package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"code.dltmdcks98.local/mattermost-plugin/hermes/internal/command"
	"code.dltmdcks98.local/mattermost-plugin/hermes/internal/gateway"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/pkg/errors"
)

const pluginCommandTrigger = "hermes"

type configuration struct {
	HermesAgentAPIURL string
	HermesAgentAPIKey string
	HermesAgentModel  string
	AllowedUserIDs    string
}

type Plugin struct {
	plugin.MattermostPlugin

	configurationLock sync.RWMutex
	configuration     *configuration
}

func (p *Plugin) OnActivate() error {
	if err := p.API.RegisterCommand(buildHermesCommand()); err != nil {
		return errors.Wrap(err, "failed to register /hermes command")
	}
	return nil
}

func (p *Plugin) OnConfigurationChange() error {
	next := new(configuration)
	if err := p.API.LoadPluginConfiguration(next); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}
	p.setConfiguration(next)
	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	cfg := p.getConfiguration()
	req := commandRequestFromMattermostArgs(args)
	allowedUsers := parseCSV(cfg.AllowedUserIDs)

	if !isUserAllowed(req.UserID, allowedUsers) {
		return commandResponseToMattermostResponse(command.ResponseEphemeral, "권한이 없습니다. Hermes slash 명령을 실행할 수 없습니다."), nil
	}
	if strings.TrimSpace(req.Text) == "" {
		return commandResponseToMattermostResponse(command.ResponseEphemeral, command.UsageText), nil
	}

	// Mattermost slash-command callbacks must return quickly. A full Hermes turn can
	// take longer than Mattermost's command timeout and otherwise surfaces to users
	// as a no-response error. Run the agent call asynchronously and deliver the
	// result as an ephemeral post in the originating channel/thread.
	go p.executeAndSendEphemeral(cfg, allowedUsers, req)

	return commandResponseToMattermostResponse(command.ResponseEphemeral, "Hermes 명령을 접수했습니다. 완료되면 이 채널에 비공개 응답으로 전달합니다."), nil
}

func (p *Plugin) executeAndSendEphemeral(cfg *configuration, allowedUsers []string, req command.SlashRequest) {
	gw := gateway.NewClient(gateway.Config{BaseURL: cfg.HermesAgentAPIURL, Token: cfg.HermesAgentAPIKey, Model: cfg.HermesAgentModel}, nil)
	handler := command.NewHandler(command.Config{AllowedUsers: allowedUsers}, gw)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	response := handler.Handle(ctx, req)
	text := strings.TrimSpace(response.Text)
	if text == "" {
		text = "Hermes 명령이 완료되었습니다."
	}
	p.API.SendEphemeralPost(req.UserID, &model.Post{ChannelId: req.ChannelID, RootId: req.RootID, Message: text})
}

func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()
	if p.configuration == nil {
		return &configuration{}
	}
	return p.configuration
}

func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()
	p.configuration = configuration
}

func buildHermesCommand() *model.Command {
	return &model.Command{
		Trigger:          pluginCommandTrigger,
		AutoComplete:     true,
		AutoCompleteDesc: "Hermes Agent API 명령 실행",
		AutoCompleteHint: "<명령어 또는 질문>",
		DisplayName:      "Hermes",
		Description:      "Hermes Agent API를 Mattermost에서 실행합니다.",
	}
}

func commandRequestFromMattermostArgs(args *model.CommandArgs) command.SlashRequest {
	text := strings.TrimSpace(args.Command)
	text = strings.TrimSpace(strings.TrimPrefix(text, "/"+pluginCommandTrigger))
	return command.SlashRequest{
		UserID:    args.UserId,
		ChannelID: args.ChannelId,
		RootID:    args.RootId,
		Text:      text,
	}
}

func commandResponseToMattermostResponse(responseType, text string) *model.CommandResponse {
	if responseType == "" {
		responseType = model.CommandResponseTypeEphemeral
	}
	return &model.CommandResponse{ResponseType: responseType, Text: text}
}

func isUserAllowed(userID string, allowedUsers []string) bool {
	if len(allowedUsers) == 0 {
		return true
	}
	for _, allowed := range allowedUsers {
		if allowed == userID {
			return true
		}
	}
	return false
}

func parseCSV(value string) []string {
	fields := strings.Split(value, ",")
	parsed := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			parsed = append(parsed, field)
		}
	}
	return parsed
}
