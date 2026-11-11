package service

import (
	"context"
	"fmt"
	"os"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/service/providers"
	"github.com/PDOK/uptime-operator/internal/service/providers/betterstack"
	"github.com/PDOK/uptime-operator/internal/service/providers/pingdom"
	"github.com/PDOK/uptime-operator/internal/util"
)

type UptimeProvider interface {
	// CreateOrUpdateCheck create the given check with the uptime monitoring
	// provider, or update an existing check. Needs to be idempotent!
	CreateOrUpdateCheck(ctx context.Context, check model.UptimeCheck) error

	// DeleteCheck deletes the given check from the uptime monitoring provider
	DeleteCheck(ctx context.Context, checkID string) error
}

func GetUptimeProviderSettings(providerFlags util.ProviderFlags) (providers.UptimeProviderID, any) {
	var uptimeProviderSettings any
	uptimeProviderID := providers.UptimeProviderID(providerFlags.UptimeProvider)

	// Optional provider specific flag handling
	switch uptimeProviderID { //nolint: exhaustive // no need to handle mock provider
	case providers.ProviderPingdom:
		alertUserIDs, err := util.StringsToInts(providerFlags.PingdomAlertUserIDs)
		if err != nil {
			fmt.Printf("Unable to parse 'pingdom-alert-user-ids' flag: %s", err)
			os.Exit(1)
		}
		alertIntegrationIDs, err := util.StringsToInts(providerFlags.PingdomAlertIntegrationIDs)
		if err != nil {
			fmt.Printf("Unable to parse 'pingdom-alert-integration-ids' flag: %s", err)
			os.Exit(1)
		}
		uptimeProviderSettings = pingdom.Settings{
			APIToken:       providerFlags.PingdomAPIToken,
			UserIDs:        alertUserIDs,
			IntegrationIDs: alertIntegrationIDs,
		}
	case providers.ProviderBetterStack:
		uptimeProviderSettings = betterstack.Settings{
			APIToken: providerFlags.BetterstackAPIToken,
		}
	}

	return uptimeProviderID, uptimeProviderSettings
}
