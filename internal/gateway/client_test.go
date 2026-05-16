package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.dltmdcks98.local/mattermost-hermes-slash-plugin/internal/command"
)

func TestClientPostsCompressCommandWithContext(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"압축 완료"}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, Token: "secret-token"}, server.Client())
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
	if gotPath != "/api/v1/mattermost/commands/compress" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	want := map[string]string{"user_id": "user-a", "channel_id": "channel-a", "root_id": "root-a"}
	for k, v := range want {
		if gotBody[k] != v {
			t.Fatalf("body[%s] = %q, want %q; body=%#v", k, gotBody[k], v, gotBody)
		}
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
}
