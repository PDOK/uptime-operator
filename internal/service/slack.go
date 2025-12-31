package service

import (
	"context"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/slack-go/slack"
	"golang.org/x/time/rate"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	nrOfMessagesPerSec   = 1
	nrOfMessagesPerBurst = 10
)

type Slack struct {
	webhookURL string
	channelID  string

	rateLimit *rate.Limiter
}

func NewSlack(webhookURL, channelID string) *Slack {
	return &Slack{
		webhookURL: webhookURL,
		channelID:  channelID,

		// see https://api.slack.com/apis/rate-limits
		rateLimit: rate.NewLimiter(nrOfMessagesPerSec, nrOfMessagesPerBurst),
	}
}

func (s *Slack) Send(ctx context.Context, message string) {
	err := s.rateLimit.Wait(ctx)
	if err != nil {
		log.FromContext(ctx).Error(err, "failed waiting for slack rate limit")
	}
	err = slack.PostWebhook(s.webhookURL, &slack.WebhookMessage{
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
