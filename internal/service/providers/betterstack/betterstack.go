package betterstack

import (
	"context"
	"encoding/json"
	"fmt"
	classiclog "log"
	"net/http"
	"strconv"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
	p "github.com/PDOK/uptime-operator/internal/service/providers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const betterStackBaseURL = "https://uptime.betterstack.com"

type httpClient struct {
	client   *http.Client
	settings Settings
}

func (h httpClient) call(req *http.Request) (*http.Response, error) {
	req.Header.Set(p.HeaderAuthorization, "Bearer "+h.settings.APIToken)
	req.Header.Set(p.HeaderAccept, "application/json")
	req.Header.Set(p.HeaderContentType, "application/json")
	return h.client.Do(req)
}

func (h httpClient) get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(p.HeaderAuthorization, "Bearer "+h.settings.APIToken)
	req.Header.Set(p.HeaderAccept, "application/json")
	req.Header.Set(p.HeaderContentType, "application/json")
	return h.client.Do(req)
}

type Settings struct {
	APIToken string
}

type BetterStack struct {
	httpClient httpClient
}

// New creates a BetterStack
func New(settings Settings) *BetterStack {
	if settings.APIToken == "" {
		classiclog.Fatal("Better Stack API token is not provided")
	}
	return &BetterStack{
		httpClient: httpClient{
			client:   &http.Client{Timeout: time.Duration(5) * time.Minute},
			settings: settings,
		},
	}
}

// CreateOrUpdateCheck create the given check with Better Stack, or update an existing check. Needs to be idempotent!
func (b *BetterStack) CreateOrUpdateCheck(ctx context.Context, check model.UptimeCheck) (err error) {
	existingCheckID, err := b.findCheck(ctx, check)
	if err != nil {
		return err
	}
	if existingCheckID == p.CheckNotFound {
		log.FromContext(ctx).Info("creating new check", "check", check)
	} else {
		log.FromContext(ctx).Info("updating existing check", "check", check)
	}
	return err
}

// DeleteCheck deletes the given check from Better Stack
func (b *BetterStack) DeleteCheck(ctx context.Context, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("deleting check", "check", check)
	// TODO
	return nil
}

func (b *BetterStack) findCheck(ctx context.Context, check model.UptimeCheck) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, betterStackBaseURL+"/api/v3/metadata", nil)
	if err != nil {
		return p.CheckNotFound, err
	}
	resp, err := b.httpClient.call(req)
	if err != nil {
		return p.CheckNotFound, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return p.CheckNotFound, fmt.Errorf("got status %d, expected %d when listing metadata", resp.StatusCode, http.StatusOK)
	}
	var metadataResponse *MetadataListResponse
	err = json.NewDecoder(resp.Body).Decode(&metadataResponse)
	if err != nil {
		return p.CheckNotFound, err
	}
	for {
		for _, md := range metadataResponse.Data {
			if md.Attributes.Key == check.ID {
				monitorID, err := strconv.ParseInt(md.Attributes.OwnerID, 10, 64)
				if err != nil {
					return p.CheckNotFound, fmt.Errorf("failed to parse monitor ID %s to integer", md.Attributes.OwnerID)
				}
				return monitorID, nil
			}
		}
		if !metadataResponse.HasNext() {
			break
		}
		metadataResponse, err = metadataResponse.Next(b.httpClient)
		if err != nil {
			return p.CheckNotFound, err
		}
	}
	return p.CheckNotFound, nil
}
