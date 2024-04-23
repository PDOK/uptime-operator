package service

import (
	"context"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/slack-go/slack"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Slack struct {
	client    *slack.Client
	channelID string
}

func NewSlack(token, channelID string) *Slack {
	return &Slack{
		client:    slack.New(token),
		channelID: channelID,
	}
}

func (s *Slack) Send(ctx context.Context, message string) {
	channelID, timestamp, err := s.client.PostMessageContext(ctx, s.channelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionUsername(model.OperatorName),
		slack.MsgOptionIconEmoji(":up:"),
	)
	if err != nil {
		log.FromContext(ctx).Error(err, "failed to post Slack message",
			"message", message, "channel", channelID, "timestamp", timestamp)
	}
}
