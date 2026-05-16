package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"code.dltmdcks98.local/mattermost-hermes-slash-plugin/internal/command"
)

type Config struct {
	BaseURL string
	Token   string
}

type Client struct {
	config Config
	http   *http.Client
}

func NewClient(config Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{config: config, http: httpClient}
}

func (c *Client) Execute(ctx context.Context, cmd command.GatewayCommand) (command.GatewayResult, error) {
	if cmd.Name != "compress" {
		return command.GatewayResult{}, fmt.Errorf("unsupported gateway command: %s", cmd.Name)
	}

	body := map[string]string{
		"user_id":    cmd.UserID,
		"channel_id": cmd.ChannelID,
		"root_id":    cmd.RootID,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return command.GatewayResult{}, err
	}

	endpoint := strings.TrimRight(c.config.BaseURL, "/") + "/api/v1/mattermost/commands/compress"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return command.GatewayResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return command.GatewayResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return command.GatewayResult{}, fmt.Errorf("hermes gateway returned status %d", resp.StatusCode)
	}

	var result command.GatewayResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return command.GatewayResult{}, err
	}
	return result, nil
}
