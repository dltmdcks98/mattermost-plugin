package main

import (
	"strings"
	"testing"

	"code.dltmdcks98.local/mattermost-plugin/hermes/internal/command"
	"github.com/mattermost/mattermost/server/public/model"
)

func TestCommandRequestFromMattermostArgsStripsTrigger(t *testing.T) {
	req := commandRequestFromMattermostArgs(&model.CommandArgs{
		UserId:    "user-a",
		ChannelId: "channel-a",
		RootId:    "root-a",
		Command:   "/hermes   compress   ",
	})

	if req.UserID != "user-a" || req.ChannelID != "channel-a" || req.RootID != "root-a" {
		t.Fatalf("request context = %#v", req)
	}
	if req.Text != "compress" {
		t.Fatalf("Text = %q, want compress", req.Text)
	}
}

func TestCommandResponseToMattermostResponse(t *testing.T) {
	got := commandResponseToMattermostResponse("ephemeral", "완료")

	if got.ResponseType != model.CommandResponseTypeEphemeral {
		t.Fatalf("ResponseType = %q", got.ResponseType)
	}
	if got.Text != "완료" {
		t.Fatalf("Text = %q", got.Text)
	}
}

func TestBuildHermesCommand(t *testing.T) {
	cmd := buildHermesCommand()

	if cmd.Trigger != "hermes" {
		t.Fatalf("Trigger = %q", cmd.Trigger)
	}
	if !cmd.AutoComplete {
		t.Fatal("AutoComplete = false, want true")
	}
	if cmd.AutoCompleteHint != "<명령어 또는 질문>" {
		t.Fatalf("AutoCompleteHint = %q", cmd.AutoCompleteHint)
	}
}

func TestExecuteCommandReturnsUsageSynchronously(t *testing.T) {
	plugin := &Plugin{}
	plugin.setConfiguration(&configuration{})

	got, appErr := plugin.ExecuteCommand(nil, &model.CommandArgs{UserId: "user-a", ChannelId: "channel-a", Command: "/hermes   "})
	if appErr != nil {
		t.Fatalf("ExecuteCommand appErr = %v", appErr)
	}
	if got.ResponseType != model.CommandResponseTypeEphemeral {
		t.Fatalf("ResponseType = %q", got.ResponseType)
	}
	if !strings.Contains(got.Text, command.UsageText) {
		t.Fatalf("Text = %q, want usage", got.Text)
	}
	if got.Props == nil {
		t.Fatal("Props = nil, want card attachments")
	}
}

func TestIsUserAllowed(t *testing.T) {
	if !isUserAllowed("user-a", nil) {
		t.Fatal("empty allowlist should allow all users")
	}
	if !isUserAllowed("user-a", []string{"user-b", "user-a"}) {
		t.Fatal("matching user should be allowed")
	}
	if isUserAllowed("user-a", []string{"user-b"}) {
		t.Fatal("non-matching user should be denied")
	}
}
