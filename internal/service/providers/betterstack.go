package providers

import (
	"context"
	"encoding/json"
	"fmt"
	classiclog "log"
	"net/http"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const betterStackBaseURL = "https://uptime.betterstack.com"

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

func (m *BetterStackUptimeProvider) findCheck(ctx context.Context, check model.UptimeCheck) (int64, error) {
	result := checkNotFound

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v3/metadata", betterStackBaseURL), nil)
	if err != nil {
		return result, err
	}
	req.Header.Add(headerAccept, "application/json")
	resp, err := m.execRequest(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("got status %d, expected HTTP OK when listing metadata", resp.StatusCode)
	}

	metadataResponse := make(map[string]any)
	err = json.NewDecoder(resp.Body).Decode(&metadataResponse)
	if err != nil {
		return result, err
	}

	metadata := metadataResponse["data"].([]any)
	for _, metadataEntry := range metadata {
		// TODO find pointer to monitor
		println(metadataEntry)
	}
	return result, nil
}

func (m *BetterStackUptimeProvider) execRequest(req *http.Request) (*http.Response, error) {
	req.Header.Add(headerAuthorization, "Bearer "+m.settings.APIToken)
	return m.httpClient.Do(req)
}
