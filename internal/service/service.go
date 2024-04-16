package service

import (
	"context"
	"fmt"

	m "github.com/PDOK/uptime-operator/internal/model"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type UptimeCheckService struct {
	provider UptimeProvider
	slack    *Slack
}

func New(provider string, slackToken string, slackChannel string) *UptimeCheckService {
	var p UptimeProvider
	switch provider {
	case "mock":
		p = NewMockUptimeProvider()
		// TODO add new case(s) for actual uptime monitoring SaaS providers
	}

	var slack *Slack
	if slackToken != "" && slackChannel != "" {
		slack = NewSlack(slackToken, slackChannel)
	}

	return &UptimeCheckService{
		slack:    slack,
		provider: p,
	}
}

func (r *UptimeCheckService) Mutate(ctx context.Context, mutation m.Mutation, annotations map[string]string) {
	check := m.NewUptimeCheck(annotations)
	if check != nil {
		switch mutation {
		case m.CreateOrUpdate:
			err := r.provider.CreateOrUpdateCheck(*check)
			r.logMutation(ctx, err, mutation, check)
		case m.Delete:
			err := r.provider.DeleteCheck(*check)
			r.logMutation(ctx, err, mutation, check)
		}
	}
}

func (r *UptimeCheckService) logMutation(ctx context.Context, err error, mutation m.Mutation, check *m.UptimeCheck) {
	if err != nil {
		msg := fmt.Sprintf("%s of uptime check '%s' (id: %s) failed", string(mutation), check.Name, check.ID)
		log.FromContext(ctx).Error(err, msg, "check", check)
		if r.slack != nil {
			r.slack.SendSlackMessage(ctx, ":large_red_square: "+msg)
		}

	} else {
		msg := fmt.Sprintf("%s of uptime check '%s' (id: %s) succeeded", string(mutation), check.Name, check.ID)
		log.FromContext(ctx).Info(msg)
		if r.slack != nil {
			emoji := ":large_green_square: "
			if mutation == m.Delete {
				emoji = ":warning: "
			}
			r.slack.SendSlackMessage(ctx, emoji+msg)
		}
	}
}
