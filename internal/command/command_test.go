package command

import (
	"context"
	"errors"
	"testing"
)

type fakeGateway struct {
	calls    []GatewayCommand
	response GatewayResult
	err      error
}

func (f *fakeGateway) Execute(ctx context.Context, cmd GatewayCommand) (GatewayResult, error) {
	f.calls = append(f.calls, cmd)
	if f.err != nil {
		return GatewayResult{}, f.err
	}
	return f.response, nil
}

func TestHandleRejectsUnauthorizedUser(t *testing.T) {
	gateway := &fakeGateway{}
	handler := NewHandler(Config{AllowedUsers: []string{"allowed-user"}}, gateway)

	resp := handler.Handle(context.Background(), SlashRequest{
		UserID:    "blocked-user",
		ChannelID: "channel-a",
		RootID:    "root-a",
		Text:      "compress",
	})

	if resp.ResponseType != ResponseEphemeral {
		t.Fatalf("ResponseType = %q, want %q", resp.ResponseType, ResponseEphemeral)
	}
	if resp.Text != "권한이 없습니다. Hermes slash 명령을 실행할 수 없습니다." {
		t.Fatalf("Text = %q", resp.Text)
	}
	if len(gateway.calls) != 0 {
		t.Fatalf("gateway was called %d times, want 0", len(gateway.calls))
	}
}

func TestHandleCompressTrimsWhitespaceAndPreservesContext(t *testing.T) {
	gateway := &fakeGateway{response: GatewayResult{Message: "압축 완료"}}
	handler := NewHandler(Config{AllowedUsers: []string{"user-a"}}, gateway)

	resp := handler.Handle(context.Background(), SlashRequest{
		UserID:    "user-a",
		ChannelID: "channel-a",
		RootID:    "thread-root",
		Text:      "   compress   ",
	})

	if resp.ResponseType != ResponseEphemeral {
		t.Fatalf("ResponseType = %q, want %q", resp.ResponseType, ResponseEphemeral)
	}
	if resp.Text != "압축 완료" {
		t.Fatalf("Text = %q", resp.Text)
	}
	if len(gateway.calls) != 1 {
		t.Fatalf("gateway calls = %d, want 1", len(gateway.calls))
	}
	got := gateway.calls[0]
	if got.Name != "compress" {
		t.Fatalf("Name = %q, want compress", got.Name)
	}
	if got.UserID != "user-a" || got.ChannelID != "channel-a" || got.RootID != "thread-root" {
		t.Fatalf("context = %#v", got)
	}
}

func TestHandleReturnsUsageForUnknownCommand(t *testing.T) {
	gateway := &fakeGateway{}
	handler := NewHandler(Config{}, gateway)

	resp := handler.Handle(context.Background(), SlashRequest{UserID: "user-a", Text: "unknown"})

	if resp.ResponseType != ResponseEphemeral {
		t.Fatalf("ResponseType = %q, want %q", resp.ResponseType, ResponseEphemeral)
	}
	if resp.Text != UsageText {
		t.Fatalf("Text = %q, want usage", resp.Text)
	}
	if len(gateway.calls) != 0 {
		t.Fatalf("gateway was called %d times, want 0", len(gateway.calls))
	}
}

func TestHandleReturnsSafeFailureWhenGatewayFails(t *testing.T) {
	gateway := &fakeGateway{err: errors.New("secret token abc123 leaked detail")}
	handler := NewHandler(Config{}, gateway)

	resp := handler.Handle(context.Background(), SlashRequest{UserID: "user-a", Text: "compress"})

	if resp.ResponseType != ResponseEphemeral {
		t.Fatalf("ResponseType = %q, want %q", resp.ResponseType, ResponseEphemeral)
	}
	if resp.Text != "Hermes 명령 실행에 실패했습니다. 잠시 후 다시 시도해주세요." {
		t.Fatalf("Text = %q", resp.Text)
	}
}
