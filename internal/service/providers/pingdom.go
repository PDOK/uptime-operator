package providers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
)

type PingdomSettings struct {
	APIToken            string
	AlertUserIDs        []string
	AlertIntegrationIDs []string
}

type PingdomUptimeProvider struct {
	settings   PingdomSettings
	httpClient *http.Client
}

func NewPingdomUptimeProvider(settings PingdomSettings) *PingdomUptimeProvider {
	if settings.APIToken == "" {
		log.Fatal("Pingdom API token is not provided")
	}
	return &PingdomUptimeProvider{
		settings:   settings,
		httpClient: &http.Client{Timeout: time.Duration(30) * time.Second},
	}
}

func (m *PingdomUptimeProvider) HasCheck(_ model.UptimeCheck) bool {
	return true
}

func (m *PingdomUptimeProvider) CreateOrUpdateCheck(_ model.UptimeCheck) error {
	req, err := http.NewRequest(http.MethodGet, "https://api.pingdom.com/api/3.1/checks?include_tags=true", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", m.settings.APIToken))
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("got status %d, expected HTTP OK when listing existing checks", resp.StatusCode)
	}
	print(io.ReadAll(resp.Body))
	return nil
}

func (m *PingdomUptimeProvider) DeleteCheck(_ model.UptimeCheck) error {
	return nil
}
