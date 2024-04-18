package service

import (
	"github.com/PDOK/uptime-operator/internal/model"
)

type UptimeProvider interface {
	// HasCheck true when the check with the given ID exists, false otherwise
	HasCheck(check model.UptimeCheck) bool

	// CreateOrUpdateCheck create the given check with the uptime monitoring
	// provider, or update an existing check. Needs to be idempotent!
	CreateOrUpdateCheck(check model.UptimeCheck) error

	// DeleteCheck deletes the given check with from the uptime monitoring provider
	DeleteCheck(check model.UptimeCheck) error
}
