package providers

import (
	"context"
	classiclog "log"
	"net/http"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const betterstackBaseURL = "https://uptime.betterstack.com/api"

type BetterStackSettings struct {
	APIToken string
}

type BetterStackUptimeProvider struct {
	settings   BetterStackSettings
	httpClient *http.Client
}

// NewBetterStackProvider creates a BetterStackUptimeProvider
func NewBetterStackProvider(settings BetterStackSettings) *BetterStackUptimeProvider {
	if settings.APIToken == "" {
		classiclog.Fatal("Better Stack API token is not provided")
	}
	return &BetterStackUptimeProvider{
		settings:   settings,
		httpClient: &http.Client{Timeout: time.Duration(5) * time.Minute},
	}
}

// CreateOrUpdateCheck create the given check with Better Stack, or update an existing check. Needs to be idempotent!
func (m *BetterStackUptimeProvider) CreateOrUpdateCheck(ctx context.Context, check model.UptimeCheck) (err error) {
	// TODO
	return err
}

// DeleteCheck deletes the given check from Better Stack
func (m *BetterStackUptimeProvider) DeleteCheck(ctx context.Context, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("deleting check", "check", check)
	// TODO
	return nil
}
