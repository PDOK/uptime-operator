package provider

import (
	"github.com/PDOK/uptime-operator/internal/model"
)

type UptimeProvider interface {
	HasCheck(check model.UptimeCheck) bool
	CreateOrUpdateCheck(check model.UptimeCheck) error
	DeleteCheck(check model.UptimeCheck) error
}
