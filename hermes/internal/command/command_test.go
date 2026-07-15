package command

import (
	"context"
	"errors"
	"strings"
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

func TestHandleForwardsArbitraryHermesCommandText(t *testing.T) {
	gateway := &fakeGateway{response: GatewayResult{Message: "실행 완료"}}
	handler := NewHandler(Config{AllowedUsers: []string{"user-a"}}, gateway)

	resp := handler.Handle(context.Background(), SlashRequest{
		UserID:    "user-a",
		ChannelID: "channel-a",
		RootID:    "thread-root",
		Text:      "  오늘 작업 요약해줘  ",
	})

	if resp.ResponseType != ResponseEphemeral {
		t.Fatalf("ResponseType = %q, want %q", resp.ResponseType, ResponseEphemeral)
	}
	if resp.Text != "실행 완료" {
		t.Fatalf("Text = %q", resp.Text)
	}
	if len(gateway.calls) != 1 {
		t.Fatalf("gateway calls = %d, want 1", len(gateway.calls))
	}
	got := gateway.calls[0]
	if got.Name != "오늘 작업 요약해줘" {
		t.Fatalf("Name = %q, want arbitrary command text", got.Name)
	}
}

func TestHandleReturnsUsageForEmptyCommand(t *testing.T) {
	gateway := &fakeGateway{}
	handler := NewHandler(Config{}, gateway)

	resp := handler.Handle(context.Background(), SlashRequest{UserID: "user-a", Text: "   "})

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

func TestHandleBlocksGatewayOnlyApprovalCommands(t *testing.T) {
	gateway := &fakeGateway{response: GatewayResult{Message: "should not be used"}}
	handler := NewHandler(Config{}, gateway)

	resp := handler.Handle(context.Background(), SlashRequest{UserID: "user-a", Text: " approve all "})

	if resp.ResponseType != ResponseEphemeral {
		t.Fatalf("ResponseType = %q, want %q", resp.ResponseType, ResponseEphemeral)
	}
	if len(gateway.calls) != 0 {
		t.Fatalf("gateway was called %d times, want 0", len(gateway.calls))
	}
	for _, want := range []string{"Gateway", "@hermes /approve", "/hermes approve"} {
		if !strings.Contains(resp.Text, want) {
			t.Fatalf("Text missing %q: %s", want, resp.Text)
		}
	}
}

func TestHandleReturnsLocalHelpStatusAndSessionGuidance(t *testing.T) {
	for _, tc := range []struct {
		text string
		want string
	}{
		{text: "help", want: "Hermes Mattermost Slash Plugin"},
		{text: "status", want: "slash plugin is active"},
		{text: "model", want: "모델 전환 안내"},
		{text: "new", want: "새 세션 안내"},
	} {
		t.Run(tc.text, func(t *testing.T) {
			gateway := &fakeGateway{response: GatewayResult{Message: "should not be used"}}
			handler := NewHandler(Config{}, gateway)

			resp := handler.Handle(context.Background(), SlashRequest{UserID: "user-a", Text: tc.text})

			if resp.ResponseType != ResponseEphemeral {
				t.Fatalf("ResponseType = %q, want %q", resp.ResponseType, ResponseEphemeral)
			}
			if len(gateway.calls) != 0 {
				t.Fatalf("gateway was called %d times, want 0", len(gateway.calls))
			}
			if !strings.Contains(resp.Text, tc.want) {
				t.Fatalf("Text missing %q: %s", tc.want, resp.Text)
			}
		})
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
