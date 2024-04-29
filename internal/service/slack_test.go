package service

import (
	"context"
	"os"
	"testing"
)

func TestSlack_Send(t *testing.T) {
	t.Skip() // only for local testing
	type slackConfig struct {
		webhookURL string
		channelID  string
	}
	tests := []struct {
		name    string
		fields  slackConfig
		message string
	}{
		{
			name: "test send to some webhook url",
			fields: slackConfig{
				webhookURL: os.Getenv("SLACK_WEBHOOK_URL"), // secret!
				channelID:  os.Getenv("SLACK_CHANNEL_ID"),
			},
			message: ":warning:\ntest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			s := &Slack{
				webhookURL: tt.fields.webhookURL,
				channelID:  tt.fields.channelID,
			}
			s.Send(context.TODO(), tt.message)
		})
	}
}
