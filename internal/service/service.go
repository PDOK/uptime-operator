package service

import (
	"context"
	"fmt"

	m "github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/service/providers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type UptimeCheckService struct {
	provider UptimeProvider
	slack    *Slack
}

func New(provider string, slackToken string, slackChannel string) *UptimeCheckService {
	var p UptimeProvider
	switch provider { //nolint:gocritic
	case "mock":
		p = providers.NewMockUptimeProvider()
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
		if mutation == m.CreateOrUpdate {
			err := r.provider.CreateOrUpdateCheck(*check)
			r.logMutation(ctx, err, mutation, check)
		} else if mutation == m.Delete {
			err := r.provider.DeleteCheck(*check)
			r.logMutation(ctx, err, mutation, check)
		}
	}
}

func (r *UptimeCheckService) logMutation(ctx context.Context, err error, mutation m.Mutation, check *m.UptimeCheck) {
	if err != nil {
		msg := fmt.Sprintf("%s of uptime check '%s' (id: %s) failed", string(mutation), check.Name, check.ID)
		log.FromContext(ctx).Error(err, msg, "check", check)
		if r.slack == nil {
			return
		}
		r.slack.SendSlackMessage(ctx, ":large_red_square: "+msg)
	} else {
		msg := fmt.Sprintf("%s of uptime check '%s' (id: %s) succeeded", string(mutation), check.Name, check.ID)
		log.FromContext(ctx).Info(msg)
		if r.slack == nil {
			return
		}
		if mutation == m.Delete {
			r.slack.SendSlackMessage(ctx, ":warning: "+msg+". Beware: a flood of these delete message may indicate Traefik itself is down!")
		} else {
			r.slack.SendSlackMessage(ctx, ":large_green_square: "+msg)
		}
	}
}
