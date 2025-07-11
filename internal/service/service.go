package service

import (
	"context"
	"fmt"
	classiclog "log"

	m "github.com/PDOK/uptime-operator/internal/model"
	p "github.com/PDOK/uptime-operator/internal/service/providers"
	"github.com/PDOK/uptime-operator/internal/service/providers/betterstack"
	"github.com/PDOK/uptime-operator/internal/service/providers/mock"
	"github.com/PDOK/uptime-operator/internal/service/providers/pingdom"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type UptimeCheckOption func(*UptimeCheckService) *UptimeCheckService

type UptimeCheckService struct {
	provider      UptimeProvider
	slack         *Slack
	enableDeletes bool
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

func WithProviderAndSettings(provider p.UptimeProviderID, settings any) UptimeCheckOption {
	return func(service *UptimeCheckService) *UptimeCheckService {
		switch provider {
		case p.ProviderMock:
			service.provider = mock.New()
		case p.ProviderPingdom:
			service.provider = pingdom.New(settings.(pingdom.Settings))
		case p.ProviderBetterStack:
			service.provider = betterstack.New(settings.(betterstack.Settings))
		default:
			classiclog.Fatalf("unsupported provider specified: %s", provider)
		}
		return service
	}
}

func WithSlack(slackWebhookURL string, slackChannel string) UptimeCheckOption {
	return func(service *UptimeCheckService) *UptimeCheckService {
		if slackWebhookURL != "" && slackChannel != "" {
			service.slack = NewSlack(slackWebhookURL, slackChannel)
		}
		return service
	}
}

func WithDeletes(enableDeletes bool) UptimeCheckOption {
	return func(service *UptimeCheckService) *UptimeCheckService {
		service.enableDeletes = enableDeletes
		return service
	}
}

func (r *UptimeCheckService) Mutate(ctx context.Context, mutation m.Mutation, ingressName string, annotations map[string]string) {
	_, ignore := annotations[m.AnnotationIgnore]
	if ignore {
		r.logRouteIgnore(ctx, mutation, ingressName)
		return
	}
	check, err := m.NewUptimeCheck(ingressName, annotations)
	if err != nil {
		r.logAnnotationErr(ctx, err)
		return
	}
	if mutation == m.CreateOrUpdate {
		err = r.provider.CreateOrUpdateCheck(ctx, *check)
		r.logMutation(ctx, err, mutation, check)
	} else if mutation == m.Delete {
		if !r.enableDeletes {
			r.logDeleteDisabled(ctx, check)
			return
		}
		err = r.provider.DeleteCheck(ctx, *check)
		r.logMutation(ctx, err, mutation, check)
	}
}

func (r *UptimeCheckService) logDeleteDisabled(ctx context.Context, check *m.UptimeCheck) {
	msg := fmt.Sprintf("delete of uptime check '%s' (id: %s) not executed since 'enable-deletes=false'.", check.Name, check.ID)
	log.FromContext(ctx).Info(msg, "check", check)
	if r.slack == nil {
		return
	}
	r.slack.Send(ctx, ":information_source: "+msg)
}

func (r *UptimeCheckService) logRouteIgnore(ctx context.Context, mutation m.Mutation, name string) {
	msg := fmt.Sprintf("ignoring %s for ingress route %s, since this route is marked to be excluded from uptime monitoring", mutation, name)
	log.FromContext(ctx).Info(msg)
	if r.slack == nil {
		return
	}
	r.slack.Send(ctx, ":information_source: "+msg)
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
