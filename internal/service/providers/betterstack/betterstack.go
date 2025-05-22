package betterstack

import (
	"context"
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
	req.Header.Set(p.HeaderAccept, p.MediaTypeJSON)
	req.Header.Set(p.HeaderContentType, p.MediaTypeJSON)
	req.Header.Add(p.HeaderUserAgent, model.OperatorName)
	return h.client.Do(req)
}

func (h httpClient) get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(p.HeaderAuthorization, "Bearer "+h.settings.APIToken)
	req.Header.Set(p.HeaderAccept, p.MediaTypeJSON)
	req.Header.Set(p.HeaderContentType, p.MediaTypeJSON)
	req.Header.Add(p.HeaderUserAgent, model.OperatorName)
	return h.client.Do(req)
}

type Settings struct {
	APIToken string
	PageSize int
}

type BetterStack struct {
	httpClient httpClient
}

// New creates a BetterStack
func New(settings Settings) *BetterStack {
	if settings.APIToken == "" {
		classiclog.Fatal("Better Stack API token is not provided")
	}
	if settings.PageSize < 1 {
		settings.PageSize = 50 // def
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
	existingCheckID, err := b.findCheck(check)
	if err != nil {
		return err
	}
	if existingCheckID == p.CheckNotFound {
		log.FromContext(ctx).Info("creating check", "check", check)
		monitorID, err := CreateMonitor(b.httpClient, check)
		if err != nil {
			return err
		}
		if err = CreateOrUpdateMetadata(b.httpClient, check.ID, monitorID, check.Tags); err != nil {
			return err
		}
	} else {
		err = b.updateCheck(ctx, existingCheckID, check)
	}
	return err
}

// DeleteCheck deletes the given check from Better Stack
func (b *BetterStack) DeleteCheck(ctx context.Context, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("deleting check", "check", check)

	existingCheckID, err := b.findCheck(check)
	if err != nil {
		return err
	}
	if existingCheckID == p.CheckNotFound {
		log.FromContext(ctx).Info(fmt.Sprintf("check with ID '%s' is already deleted", check.ID))
		return nil
	}
	if err = DeleteMetadata(b.httpClient, check.ID, existingCheckID); err != nil {
		return err
	}
	if err = DeleteMonitor(b.httpClient, existingCheckID); err != nil {
		return err
	}
	return nil
}

func (b *BetterStack) findCheck(check model.UptimeCheck) (int64, error) {
	result := p.CheckNotFound
	metadata, err := ListMetadata(b.httpClient)
	if err != nil {
		return result, err
	}
	for {
		for _, md := range metadata.Data {
			if md.Attributes != nil && md.Attributes.Key == check.ID {
				result, err = strconv.ParseInt(md.Attributes.OwnerID, 10, 64)
				if err != nil {
					return result, fmt.Errorf("failed to parse monitor ID %s to integer", md.Attributes.OwnerID)
				}
				return result, nil
			}
		}
		if !metadata.HasNext() {
			break // exit infinite loop
		}
		metadata, err = metadata.Next(b.httpClient)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (b *BetterStack) updateCheck(ctx context.Context, existingCheckID int64, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("updating check", "check", check, "betterstack ID", existingCheckID)

	existingMonitor, err := GetMonitor(b.httpClient, existingCheckID)
	if err != nil {
		return err
	}
	if err = UpdateMonitor(b.httpClient, check, existingMonitor); err != nil {
		return err
	}
	// Update tags in particular (note: existing tags will be overwritten, old tags won't be deleted)
	if err = CreateOrUpdateMetadata(b.httpClient, check.ID, existingCheckID, check.Tags); err != nil {
		return err
	}
	return nil
}
