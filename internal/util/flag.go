package util

import (
	"flag"
	"strings"
)

type SliceFlag []string

func (sf *SliceFlag) String() string {
	return strings.Join(*sf, ",")
}

func (sf *SliceFlag) Set(value string) error {
	*sf = append(*sf, value)
	return nil
}

func GetProviderFlags() ProviderFlags {
	result := ProviderFlags{
		UptimeProvider:             "",
		PingdomAPIToken:            "",
		PingdomAlertUserIDs:        SliceFlag{},
		PingdomAlertIntegrationIDs: SliceFlag{},
		BetterstackAPIToken:        "",
	}

	flag.StringVar(&result.UptimeProvider, "uptime-provider", "mock",
		"Name of the (SaaS) uptime monitoring provider to use.")

	// Pingdom specific
	flag.StringVar(&result.PingdomAPIToken, "pingdom-api-token", "",
		"The API token to authenticate with Pingdom. Only applies when 'uptime-provider' is 'pingdom'")
	flag.Var(&result.PingdomAlertUserIDs, "pingdom-alert-user-ids",
		"One or more IDs of Pingdom users to alert. Only applies when 'uptime-provider' is 'pingdom'")
	flag.Var(&result.PingdomAlertIntegrationIDs, "pingdom-alert-integration-ids",
		"One or more IDs of Pingdom integrations (like slack channels) to alert. Only applies when 'uptime-provider' is 'pingdom'")

	// Better Stack specific
	flag.StringVar(&result.BetterstackAPIToken, "betterstack-api-token", "",
		"The API token to authenticate with Better Stack. Only applies when 'uptime-provider' is 'betterstack'")

	return result
}

type ProviderFlags struct {
	UptimeProvider             string
	PingdomAPIToken            string
	PingdomAlertUserIDs        SliceFlag
	PingdomAlertIntegrationIDs SliceFlag
	BetterstackAPIToken        string
}
