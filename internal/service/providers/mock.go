package providers

import (
	"encoding/json"
	"log"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/service"
)

type MockUptimeProvider struct {
	checks map[string]model.UptimeCheck
}

func NewMockUptimeProvider() service.UptimeProvider {
	return &MockUptimeProvider{
		checks: make(map[string]model.UptimeCheck),
	}
}

func (m *MockUptimeProvider) HasCheck(check model.UptimeCheck) bool {
	_, ok := m.checks[check.ID]
	return ok
}

func (m *MockUptimeProvider) CreateOrUpdateCheck(check model.UptimeCheck) error {
	m.checks[check.ID] = check

	checkJSON, _ := json.Marshal(check)
	log.Printf("MOCK: created or updated check %s\n", checkJSON)

	return nil
}

func (m *MockUptimeProvider) DeleteCheck(check model.UptimeCheck) error {
	delete(m.checks, check.ID)

	checkJSON, _ := json.Marshal(check)
	log.Printf("MOCK: deleted check %s\n", checkJSON)

	return nil
}
