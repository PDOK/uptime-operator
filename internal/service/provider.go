package service

import (
	"context"

	"github.com/PDOK/uptime-operator/internal/model"
)

type UptimeProvider interface {
	// CreateOrUpdateCheck create the given check with the uptime monitoring
	// provider, or update an existing check. Needs to be idempotent!
	CreateOrUpdateCheck(ctx context.Context, check model.UptimeCheck) error

	// DeleteCheck deletes the given check from the uptime monitoring provider
	DeleteCheck(ctx context.Context, check model.UptimeCheck) error
}
