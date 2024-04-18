package service

import (
	"context"
	"fmt"

	m "github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/service/providers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type UptimeCheckOption func(*UptimeCheckService) *UptimeCheckService

type UptimeCheckService struct {
	provider UptimeProvider
	slack    *Slack
}

func New(options ...UptimeCheckOption) *UptimeCheckService {
	service := &UptimeCheckService{}
	for _, option := range options {
		service = option(service)
	}
	return service
}

func WithProvider(provider UptimeProvider) UptimeCheckOption {
	return func(service *UptimeCheckService) *UptimeCheckService {
		service.provider = provider
		return service
	}
}

func WithProviderName(provider string) UptimeCheckOption {
	return func(service *UptimeCheckService) *UptimeCheckService {
		switch provider { //nolint:gocritic
		case "mock":
			service.provider = providers.NewMockUptimeProvider()
			// TODO add new case(s) for actual uptime monitoring SaaS providers
		}
		return service
	}
}

func WithSlack(slackToken string, slackChannel string) UptimeCheckOption {
	return func(service *UptimeCheckService) *UptimeCheckService {
		if slackToken != "" && slackChannel != "" {
			service.slack = NewSlack(slackToken, slackChannel)
		}
		return service
	}
}

func (r *UptimeCheckService) Mutate(ctx context.Context, mutation m.Mutation, ingressName string, annotations map[string]string) {
	check, err := m.NewUptimeCheck(ingressName, annotations)
	if err != nil {
		r.logAnnotationErr(ctx, err)
		return
	}
	if mutation == m.CreateOrUpdate {
		err = r.provider.CreateOrUpdateCheck(*check)
		r.logMutation(ctx, err, mutation, check)
	} else if mutation == m.Delete {
		err = r.provider.DeleteCheck(*check)
		r.logMutation(ctx, err, mutation, check)
	}
}

func (r *UptimeCheckService) logAnnotationErr(ctx context.Context, err error) {
	msg := fmt.Sprintf("missing or invalid uptime check annotation(s) encountered: %v", err)
	log.FromContext(ctx).Error(err, msg)
	if r.slack == nil {
		return
	}
	r.slack.Send(ctx, ":large_red_square: "+msg)
}

func (r *UptimeCheckService) logMutation(ctx context.Context, err error, mutation m.Mutation, check *m.UptimeCheck) {
	if err != nil {
		msg := fmt.Sprintf("%s of uptime check '%s' (id: %s) failed.", string(mutation), check.Name, check.ID)
		log.FromContext(ctx).Error(err, msg, "check", check)
		if r.slack == nil {
			return
		}
		r.slack.Send(ctx, ":large_red_square: "+msg)
		return
	}
	msg := fmt.Sprintf("%s of uptime check '%s' (id: %s) succeeded.", string(mutation), check.Name, check.ID)
	log.FromContext(ctx).Info(msg)
	if r.slack == nil {
		return
	}
	if mutation == m.Delete {
		r.slack.Send(ctx, ":warning: "+msg+".\n _Beware: a flood of these delete messages may indicate Traefik itself is down!_")
	} else {
		r.slack.Send(ctx, ":large_green_square: "+msg)
	}
}
