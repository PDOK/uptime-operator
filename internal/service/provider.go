package service

import (
	"github.com/PDOK/uptime-operator/internal/model"
)

type UptimeProvider interface {
	// CreateOrUpdateCheck create the given check with the uptime monitoring
	// provider, or update an existing check. Needs to be idempotent!
	CreateOrUpdateCheck(check model.UptimeCheck) error

	// DeleteCheck deletes the given check from the uptime monitoring provider
	DeleteCheck(check model.UptimeCheck) error
}
