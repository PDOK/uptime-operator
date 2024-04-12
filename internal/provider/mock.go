package provider

import (
	"encoding/json"
	"log"

	"github.com/PDOK/uptime-operator/internal/model"
)

type MockUptimeProvider struct {
	checks map[string]model.UptimeCheck
}

func NewMockUptimeProvider() UptimeProvider {
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

	checkJson, _ := json.Marshal(check)
	log.Printf("created or updated check %s\n", checkJson)

	return nil
}

func (m *MockUptimeProvider) DeleteCheck(check model.UptimeCheck) error {
	delete(m.checks, check.ID)

	checkJson, _ := json.Marshal(check)
	log.Printf("deleted check %s\n", checkJson)

	return nil
}
