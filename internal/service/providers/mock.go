package providers

import (
	"context"
	"encoding/json"

	"github.com/PDOK/uptime-operator/internal/model"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MockUptimeProvider struct {
	checks map[string]model.UptimeCheck
}

func NewMockUptimeProvider() *MockUptimeProvider {
	return &MockUptimeProvider{
		checks: make(map[string]model.UptimeCheck),
	}
}

func (m *MockUptimeProvider) CreateOrUpdateCheck(ctx context.Context, check model.UptimeCheck) error {
	m.checks[check.ID] = check

	checkJSON, _ := json.Marshal(check)
	log.FromContext(ctx).Info("MOCK: created or updated check %s\n", checkJSON)

	return nil
}

func (m *MockUptimeProvider) DeleteCheck(ctx context.Context, check model.UptimeCheck) error {
	delete(m.checks, check.ID)

	checkJSON, _ := json.Marshal(check)
	log.FromContext(ctx).Info("MOCK: deleted check %s\n", checkJSON)

	return nil
}
