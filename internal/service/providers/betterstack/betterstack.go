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

type Settings struct {
	APIToken string
	PageSize int
}

type BetterStack struct {
	client Client
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
		Client{
			httpClient: &http.Client{Timeout: time.Duration(5) * time.Minute},
			settings:   settings,
		},
	}
}

// CreateOrUpdateCheck create the given check with Better Stack, or update an existing check. Needs to be idempotent!
func (b *BetterStack) CreateOrUpdateCheck(ctx context.Context, check model.UptimeCheck) (err error) {
	existingCheckID, err := b.findCheck(check)
	if err != nil {
		return fmt.Errorf("failed to find check %s, error: %w", check.ID, err)
	}
	if existingCheckID == p.CheckNotFound { //nolint:nestif // clean enough
		log.FromContext(ctx).Info("creating check", "check", check)
		monitorID, err := b.client.createMonitor(check)
		if err != nil {
			return fmt.Errorf("failed to create monitor for check %s, error: %w", check.ID, err)
		}
		if err = b.client.createMetadata(check.ID, monitorID, check.Tags); err != nil {
			return fmt.Errorf("failed to create metadata for check %s, error: %w", check.ID, err)
		}
	} else {
		log.FromContext(ctx).Info("updating check", "check", check, "betterstack ID", existingCheckID)
		existingMonitor, err := b.client.getMonitor(existingCheckID)
		if err != nil {
			return fmt.Errorf("failed to get monitor for check %s, error: %w", check.ID, err)
		}
		if err = b.client.updateMonitor(check, existingMonitor); err != nil {
			return fmt.Errorf("failed to update monitor for check %s (betterstack ID: %d), "+
				"error: %w", check.ID, existingCheckID, err)
		}
		if err = b.client.updateMetadata(check.ID, existingCheckID, check.Tags); err != nil {
			return fmt.Errorf("failed to update metdata for check %s (betterstack ID: %d), "+
				"error: %w", check.ID, existingCheckID, err)
		}
	}
	return err
}

// DeleteCheck deletes the given check from Better Stack
func (b *BetterStack) DeleteCheck(ctx context.Context, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("deleting check", "check", check)

	existingCheckID, err := b.findCheck(check)
	if err != nil {
		return fmt.Errorf("failed to find check %s, error: %w", check.ID, err)
	}
	if existingCheckID == p.CheckNotFound {
		log.FromContext(ctx).Info(fmt.Sprintf("check with ID '%s' is already deleted", check.ID))
		return nil
	}
	if err = b.client.deleteMetadata(check.ID, existingCheckID); err != nil {
		return fmt.Errorf("failed to delete metadata for check %s (betterstack ID: %d), "+
			"error: %w", check.ID, existingCheckID, err)
	}
	if err = b.client.deleteMonitor(existingCheckID); err != nil {
		return fmt.Errorf("failed to delete monitor for check %s (betterstack ID: %d), "+
			"error: %w", check.ID, existingCheckID, err)
	}
	return nil
}

func (b *BetterStack) findCheck(check model.UptimeCheck) (int64, error) {
	result := p.CheckNotFound
	metadata, err := b.client.listMetadata()
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
		metadata, err = metadata.Next(b.client)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}
