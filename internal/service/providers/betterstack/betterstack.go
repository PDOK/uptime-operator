package betterstack

import (
	"context"
	"encoding/json"
	"fmt"
	classiclog "log"
	"net/http"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/service/providers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const betterStackBaseURL = "https://uptime.betterstack.com"

type Settings struct {
	APIToken string
}

type BetterStack struct {
	settings   Settings
	httpClient *http.Client
}

// New creates a BetterStack
func New(settings Settings) *BetterStack {
	if settings.APIToken == "" {
		classiclog.Fatal("Better Stack API token is not provided")
	}
	return &BetterStack{
		settings:   settings,
		httpClient: &http.Client{Timeout: time.Duration(5) * time.Minute},
	}
}

// CreateOrUpdateCheck create the given check with Better Stack, or update an existing check. Needs to be idempotent!
func (b *BetterStack) CreateOrUpdateCheck(_ context.Context, _ model.UptimeCheck) (err error) {
	// TODO
	return err
}

// DeleteCheck deletes the given check from Better Stack
func (b *BetterStack) DeleteCheck(ctx context.Context, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("deleting check", "check", check)
	// TODO
	return nil
}

func (b *BetterStack) findCheck(ctx context.Context, _ model.UptimeCheck) (int64, error) {
	result := providers.CheckNotFound

	req, err := http.NewRequest(http.MethodGet, betterStackBaseURL+"/api/v3/metadata", nil)
	if err != nil {
		return result, err
	}
	req.Header.Add(providers.HeaderAccept, "application/json")
	resp, err := b.execRequest(req)
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
		log.FromContext(ctx).Info("test", "metadataEntry", metadataEntry)
	}
	return result, nil
}

func (b *BetterStack) execRequest(req *http.Request) (*http.Response, error) {
	req.Header.Add(providers.HeaderAuthorization, "Bearer "+b.settings.APIToken)
	return b.httpClient.Do(req)
}
