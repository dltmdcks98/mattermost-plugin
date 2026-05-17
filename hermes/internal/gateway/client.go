package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"code.dltmdcks98.local/mattermost-plugin/hermes/internal/command"
)

const defaultHermesModel = "hermes-agent"

// Config points the plugin at Hermes Agent's OpenAI-compatible API server.
// BaseURL may be either the server root (http://host:8642) or the v1 base
// (http://host:8642/v1). Token is sent as a bearer API_SERVER_KEY.
type Config struct {
	BaseURL string
	Token   string
	Model   string
}

type Client struct {
	config Config
	http   *http.Client
}

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func NewClient(config Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{config: config, http: httpClient}
}

func (c *Client) Execute(ctx context.Context, cmd command.GatewayCommand) (command.GatewayResult, error) {
	prompt := strings.TrimSpace(cmd.Name)
	if prompt == "" {
		return command.GatewayResult{}, fmt.Errorf("empty hermes command")
	}

	payload, err := json.Marshal(chatCompletionRequest{
		Model:  c.modelName(),
		Stream: false,
		Messages: []chatMessage{
			{
				Role: "system",
				Content: strings.TrimSpace(`You are Hermes Agent responding to a Mattermost slash command.
Return a concise Korean Mattermost Markdown response.
Do not expose secrets. If tool access is unavailable, explain the exact missing capability briefly.`),
			},
			{
				Role:    "user",
				Content: buildCommandPrompt(cmd, prompt),
			},
		},
	})
	if err != nil {
		return command.GatewayResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatCompletionsEndpoint(), bytes.NewReader(payload))
	if err != nil {
		return command.GatewayResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hermes-Session-Key", sessionKey(cmd))
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return command.GatewayResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return command.GatewayResult{}, fmt.Errorf("hermes agent API returned status %d", resp.StatusCode)
	}

	var apiResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return command.GatewayResult{}, err
	}
	if len(apiResp.Choices) == 0 {
		return command.GatewayResult{}, fmt.Errorf("hermes agent API returned no choices")
	}
	message := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	if message == "" {
		return command.GatewayResult{}, fmt.Errorf("hermes agent API returned an empty message")
	}
	return command.GatewayResult{Message: message}, nil
}

func (c *Client) modelName() string {
	model := strings.TrimSpace(c.config.Model)
	if model == "" {
		return defaultHermesModel
	}
	return model
}

func (c *Client) chatCompletionsEndpoint() string {
	base := strings.TrimRight(strings.TrimSpace(c.config.BaseURL), "/")
	if base == "" {
		base = "http://127.0.0.1:8642"
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/chat/completions"
	}
	return base + "/v1/chat/completions"
}

func sessionKey(cmd command.GatewayCommand) string {
	channelID := strings.TrimSpace(cmd.ChannelID)
	rootID := strings.TrimSpace(cmd.RootID)
	if rootID == "" {
		rootID = strings.TrimSpace(cmd.UserID)
	}
	if channelID == "" {
		channelID = "unknown-channel"
	}
	if rootID == "" {
		rootID = "unknown-root"
	}
	return "mattermost:" + channelID + ":" + rootID
}

func buildCommandPrompt(cmd command.GatewayCommand, prompt string) string {
	rootID := strings.TrimSpace(cmd.RootID)
	if rootID == "" {
		rootID = "현재 채널의 최근 대화"
	}
	return fmt.Sprintf(strings.TrimSpace(`Mattermost slash command /hermes %s was invoked.

Context identifiers:
- user_id: %s
- channel_id: %s
- root_id/thread_id: %s

Execute the user's Mattermost slash command request as Hermes Agent. Use available tools when needed and allowed. If direct Mattermost retrieval is not available in this API-server toolset, say that API access is connected but Mattermost read tooling is not available yet, and include these identifiers for follow-up.`), prompt, cmd.UserID, cmd.ChannelID, rootID)
}
