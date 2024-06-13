package service

import (
	"context"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/slack-go/slack"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Slack struct {
	webhookURL string
	channelID  string
}

func NewSlack(webhookURL, channelID string) *Slack {
	return &Slack{
		webhookURL: webhookURL,
		channelID:  channelID,
	}
}

func (s *Slack) Send(ctx context.Context, message string) {
	err := slack.PostWebhook(s.webhookURL, &slack.WebhookMessage{
		Channel:   s.channelID,
		Text:      message,
		Username:  model.OperatorName,
		IconEmoji: ":up:",
	})
	if err != nil {
		log.FromContext(ctx).Error(err, "failed to post Slack message",
			"message", message, "channel", s.channelID)
	}
}
