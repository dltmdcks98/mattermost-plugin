package main

import (
	"testing"

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
	if cmd.AutoCompleteHint != "compress" {
		t.Fatalf("AutoCompleteHint = %q", cmd.AutoCompleteHint)
	}
}
