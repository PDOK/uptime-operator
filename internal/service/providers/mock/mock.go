package mock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PDOK/uptime-operator/internal/model"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Mock struct {
	checks map[string]model.UptimeCheck
}

func New() *Mock {
	return &Mock{
		checks: make(map[string]model.UptimeCheck),
	}
}

func (m *Mock) CreateOrUpdateCheck(ctx context.Context, check model.UptimeCheck) error {
	m.checks[check.ID] = check

	checkJSON, _ := json.Marshal(check)
	log.FromContext(ctx).Info(fmt.Sprintf("MOCK: created or updated check %s\n", checkJSON))

	return nil
}

func (m *Mock) DeleteCheck(ctx context.Context, check model.UptimeCheck) error {
	delete(m.checks, check.ID)

	checkJSON, _ := json.Marshal(check)
	log.FromContext(ctx).Info(fmt.Sprintf("MOCK: deleted check %s\n", checkJSON))

	return nil
}
