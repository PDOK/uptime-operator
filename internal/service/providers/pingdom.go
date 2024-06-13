package providers

import (
	"log"
	"net/http"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
)

type PingdomSettings struct {
	ApiToken            string
	AlertUserIds        []string
	AlertIntegrationIds []string
}

type PingdomUptimeProvider struct {
	settings   PingdomSettings
	httpClient *http.Client
}

func NewPingdomUptimeProvider(settings PingdomSettings) *PingdomUptimeProvider {
	if settings.ApiToken == "" {
		log.Fatal("Pingdom API token is not provided")
	}
	return &PingdomUptimeProvider{
		settings:   settings,
		httpClient: &http.Client{Timeout: time.Duration(30) * time.Second},
	}
}

func (m *PingdomUptimeProvider) HasCheck(check model.UptimeCheck) bool {
	return true
}

func (m *PingdomUptimeProvider) CreateOrUpdateCheck(check model.UptimeCheck) error {
	return nil
}

func (m *PingdomUptimeProvider) DeleteCheck(check model.UptimeCheck) error {
	return nil
}
