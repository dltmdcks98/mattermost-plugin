package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"code.dltmdcks98.local/mattermost-hermes-slash-plugin/internal/command"
	"code.dltmdcks98.local/mattermost-hermes-slash-plugin/internal/gateway"
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
	gw := gateway.NewClient(gateway.Config{BaseURL: cfg.HermesAgentAPIURL, Token: cfg.HermesAgentAPIKey, Model: cfg.HermesAgentModel}, nil)
	handler := command.NewHandler(command.Config{AllowedUsers: parseCSV(cfg.AllowedUserIDs)}, gw)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response := handler.Handle(ctx, commandRequestFromMattermostArgs(args))
	return commandResponseToMattermostResponse(response.ResponseType, response.Text), nil
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
		AutoCompleteHint: "compress",
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
