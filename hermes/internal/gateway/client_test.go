package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.dltmdcks98.local/mattermost-plugin/hermes/internal/command"
)

func TestClientPostsCompressCommandToHermesChatCompletionsAPI(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotSessionKey string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotSessionKey = r.Header.Get("X-Hermes-Session-Key")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"압축 완료"}}]}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL + "/v1", Token: "secret-token", Model: "hermes-agent"}, server.Client())
	result, err := client.Execute(context.Background(), command.GatewayCommand{
		Name:      "compress",
		UserID:    "user-a",
		ChannelID: "channel-a",
		RootID:    "root-a",
	})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Message != "압축 완료" {
		t.Fatalf("Message = %q", result.Message)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotSessionKey != "mattermost:channel-a:root-a" {
		t.Fatalf("X-Hermes-Session-Key = %q", gotSessionKey)
	}
	if gotBody["model"] != "hermes-agent" {
		t.Fatalf("model = %#v", gotBody["model"])
	}
	if gotBody["stream"] != false {
		t.Fatalf("stream = %#v", gotBody["stream"])
	}
	messages, ok := gotBody["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("messages = %#v", gotBody["messages"])
	}
	userMsg := messages[1].(map[string]any)
	content := userMsg["content"].(string)
	for _, want := range []string{"/hermes compress", "channel-a", "root-a", "user-a"} {
		if !strings.Contains(content, want) {
			t.Fatalf("user message missing %q: %s", want, content)
		}
	}
}

func TestClientPostsArbitraryHermesCommandText(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"실행 완료"}}]}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL}, server.Client())
	result, err := client.Execute(context.Background(), command.GatewayCommand{
		Name:      "오늘 작업 요약해줘",
		UserID:    "user-a",
		ChannelID: "channel-a",
	})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Message != "실행 완료" {
		t.Fatalf("Message = %q", result.Message)
	}
	messages := gotBody["messages"].([]any)
	userMsg := messages[1].(map[string]any)
	content := userMsg["content"].(string)
	if !strings.Contains(content, "/hermes 오늘 작업 요약해줘") {
		t.Fatalf("user message missing arbitrary slash command: %s", content)
	}
}

func TestClientAcceptsBaseURLWithoutV1Suffix(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL}, server.Client())
	_, err := client.Execute(context.Background(), command.GatewayCommand{Name: "compress"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestClientReturnsErrorForNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal secret detail", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL}, server.Client())
	_, err := client.Execute(context.Background(), command.GatewayCommand{Name: "compress"})

	if err == nil {
		t.Fatal("Execute error = nil, want error")
	}
	if strings.Contains(err.Error(), "internal secret detail") {
		t.Fatalf("error leaked response body: %v", err)
	}
}

func TestClientReturnsErrorForEmptyAssistantMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":""}}]}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL}, server.Client())
	_, err := client.Execute(context.Background(), command.GatewayCommand{Name: "compress"})
	if err == nil {
		t.Fatal("Execute error = nil, want error")
	}
}
