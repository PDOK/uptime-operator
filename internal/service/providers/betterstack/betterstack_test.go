package betterstack

import (
	"context"
	"maps"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/service/providers"
	"github.com/stretchr/testify/assert"
)

func TestConfirmationPeriodEnabled(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		expected    bool
		wantErr     bool
	}{
		{name: "DefaultEnabled", expected: true},
		{name: "ExplicitlyEnabled", annotations: map[string]string{annotationConfirmationPeriodEnabled: "true"}, expected: true},
		{name: "Disabled", annotations: map[string]string{annotationConfirmationPeriodEnabled: "false"}, expected: false},
		{name: "InvalidBoolean", annotations: map[string]string{annotationConfirmationPeriodEnabled: "invalid"}, wantErr: true},
		{name: "EmptyBoolean", annotations: map[string]string{annotationConfirmationPeriodEnabled: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := confirmationPeriodEnabled(tt.annotations)
			if (err != nil) != tt.wantErr {
				t.Fatalf("confirmationPeriodEnabled() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), annotationConfirmationPeriodEnabled) {
					t.Errorf("confirmationPeriodEnabled() error = %v, expected annotation name", err)
				}
				return
			}
			if actual != tt.expected {
				t.Errorf("confirmationPeriodEnabled() => expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

func TestCreateOrUpdateCheckInvalidConfirmationPeriod(t *testing.T) {
	m := &BetterStack{}
	check := model.UptimeCheck{
		Annotations: map[string]string{annotationConfirmationPeriodEnabled: "invalid"},
	}
	err := m.CreateOrUpdateCheck(context.Background(), check)
	if err == nil || !strings.Contains(err.Error(), annotationConfirmationPeriodEnabled) {
		t.Errorf("CreateOrUpdateCheck() error = %v, expected annotation validation error before API calls", err)
	}
}

type integrationTest struct {
	name                    string
	annotations             map[string]string
	wantErr                 bool
	wantDelete              bool
	wantConfirmationPeriod  int
	checkConfirmationPeriod bool
}

// Test against production Better Stack API. Please supply BETTERSTACK_API_TOKEN.
func TestAgainstREALBetterStackAPI_ConfirmationPeriod(t *testing.T) {
	baseAnnotations := map[string]string{
		model.AnnotationID:       "5c0e1a7e-3f4b-4d2a-9b8e-c0f14710a0a0",
		model.AnnotationName:     "UptimeOperatorBetterStackConfirmationPeriodTest",
		model.AnnotationURL:      "https://service.pdok.nl/cbs/landuse/wfs/v1_0?request=GetCapabilities&service=WFS",
		model.AnnotationInterval: "1",
	}
	annotationsForBetterStackConfirmationEnabled := func(enabled string) map[string]string {
		annotations := maps.Clone(baseAnnotations)
		annotations[annotationConfirmationPeriodEnabled] = enabled
		return annotations
	}
	tests := []integrationTest{
		{
			name:                    "Create check with confirmation period disabled",
			annotations:             annotationsForBetterStackConfirmationEnabled("false"),
			wantConfirmationPeriod:  0,
			checkConfirmationPeriod: true,
		},
		{
			name:                    "Update check with confirmation period enabled",
			annotations:             annotationsForBetterStackConfirmationEnabled("true"),
			wantConfirmationPeriod:  61,
			checkConfirmationPeriod: true,
		},
		{
			name:                    "Update check with confirmation period disabled",
			annotations:             annotationsForBetterStackConfirmationEnabled("false"),
			wantConfirmationPeriod:  0,
			checkConfirmationPeriod: true,
		},
		{
			name:        "Delete confirmation period test check",
			annotations: maps.Clone(baseAnnotations),
			wantDelete:  true,
		},
	}
	for _, tt := range tests {
		runIntegrationTest(t, tt)
	}
}

// Test against production better stack API. Please supply BETTERSTACK_API_TOKEN.
// This test creates one check, updates it and then deletes the check.
func TestAgainstREALBetterStackAPI(t *testing.T) {
	tests := []integrationTest{
		{
			name: "Create check",
			annotations: map[string]string{
				"uptime.pdok.nl/id":                                     "3w2e9d804b2cd6bf18b8c0a6e1c04e46ac62b98c",
				"uptime.pdok.nl/name":                                   "UptimeOperatorBetterStackTestCheck",
				"uptime.pdok.nl/url":                                    "https://service.pdok.nl/cbs/landuse/wfs/v1_0?request=GetCapabilities&service=WFS",
				"uptime.pdok.nl/tags":                                   "tag1, tag2, TooLongTagOvEtj8xOzZmGPJNf5ZcGikHzTjAG55xvcWVymItA0O8Us9tq6fEAfRYeN6AODj2gwRRi5l",
				"uptime.pdok.nl/request-headers":                        "key1:value1, key2:value2",
				"uptime.pdok.nl/response-check-for-string-contains":     "bla",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
		},
		{
			name: "Update check",
			annotations: map[string]string{
				"uptime.pdok.nl/id":                                     "3w2e9d804b2cd6bf18b8c0a6e1c04e46ac62b98c",
				"uptime.pdok.nl/name":                                   "UptimeOperatorBetterStackTestCheck - Updated",
				"uptime.pdok.nl/url":                                    "https://service.pdok.nl/cbs/landuse/wfs/v1_0?request=GetCapabilities&service=WFS",
				"uptime.pdok.nl/tags":                                   "tag1",
				"uptime.pdok.nl/request-headers":                        "key1:value1, key2:value2, key3:value3",
				"uptime.pdok.nl/response-check-for-string-contains":     "",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
		},
		{
			name: "Update check again (test for idempotency)",
			annotations: map[string]string{
				"uptime.pdok.nl/id":                                     "3w2e9d804b2cd6bf18b8c0a6e1c04e46ac62b98c",
				"uptime.pdok.nl/name":                                   "UptimeOperatorBetterStackTestCheck - Updated",
				"uptime.pdok.nl/url":                                    "https://service.pdok.nl/cbs/landuse/wfs/v1_0?request=GetCapabilities&service=WFS",
				"uptime.pdok.nl/tags":                                   "tag1",
				"uptime.pdok.nl/request-headers":                        "key1:value1, key2:value2, key3:value3",
				"uptime.pdok.nl/response-check-for-string-contains":     "",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
		},
		{
			name: "Delete check",
			annotations: map[string]string{
				"uptime.pdok.nl/id":                                     "3w2e9d804b2cd6bf18b8c0a6e1c04e46ac62b98c",
				"uptime.pdok.nl/name":                                   "UptimeOperatorBetterStackTestCheck - Updated",
				"uptime.pdok.nl/url":                                    "https://service.pdok.nl/cbs/landuse/wfs/v1_0?request=GetCapabilities&service=WFS",
				"uptime.pdok.nl/tags":                                   "tag1",
				"uptime.pdok.nl/request-headers":                        "key1:value1, key2:value2, key3:value3",
				"uptime.pdok.nl/response-check-for-string-contains":     "bladiebla",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
			wantDelete: true,
		},
	}
	for _, tt := range tests {
		runIntegrationTest(t, tt)
	}
}

// Test against production better stack API. Please supply BETTERSTACK_API_TOKEN.
// This test creates one check, updates it and then deletes the check.
func TestAgainstREALBetterStackAPI_WithPagination(t *testing.T) {
	tests := []integrationTest{
		{
			name: "Create check 1",
			annotations: map[string]string{
				"uptime.pdok.nl/id":   "cf67b916-b752-47f8-a7b9-16b75267f8ca",
				"uptime.pdok.nl/name": "UptimeOperatorBetterStackTestCheck_1",
				"uptime.pdok.nl/url":  "https://service.pdok.nl/cbs/landuse/wfs/v1_0?request=GetCapabilities&service=WFS",
			},
		},
		{
			name: "Create check 2",
			annotations: map[string]string{
				"uptime.pdok.nl/id":   "fd3316cc-b0fc-4304-b316-ccb0fc23046f",
				"uptime.pdok.nl/name": "UptimeOperatorBetterStackTestCheck_2",
				"uptime.pdok.nl/url":  "https://service.pdok.nl/cbs/landuse/wfs/v1_0?request=GetCapabilities&service=WFS",
			},
		},
		{
			name: "Update check 2",
			annotations: map[string]string{
				"uptime.pdok.nl/id":   "fd3316cc-b0fc-4304-b316-ccb0fc23046f",
				"uptime.pdok.nl/name": "UptimeOperatorBetterStackTestCheck_2_Updated",
				"uptime.pdok.nl/url":  "https://service.pdok.nl/cbs/landuse/wms/v1_0?request=GetCapabilities&service=WMS",
			},
		},
		{
			name: "Delete check 1",
			annotations: map[string]string{
				"uptime.pdok.nl/id":   "cf67b916-b752-47f8-a7b9-16b75267f8ca",
				"uptime.pdok.nl/name": "UptimeOperatorBetterStackTestCheck_1",
				"uptime.pdok.nl/url":  "https://service.pdok.nl/cbs/landuse/wfs/v1_0?request=GetCapabilities&service=WFS",
			},
			wantDelete: true,
		},
		{
			name: "Delete check 2",
			annotations: map[string]string{
				"uptime.pdok.nl/id":   "fd3316cc-b0fc-4304-b316-ccb0fc23046f",
				"uptime.pdok.nl/name": "UptimeOperatorBetterStackTestCheck_2",
				"uptime.pdok.nl/url":  "https://service.pdok.nl/cbs/landuse/wfs/v1_0?request=GetCapabilities&service=WFS",
			},
			wantDelete: true,
		},
	}
	for _, tt := range tests {
		runIntegrationTest(t, tt)
	}
}

func runIntegrationTest(t *testing.T, tt integrationTest) {
	if os.Getenv("BETTERSTACK_API_TOKEN") == "" {
		t.Skip("skipping test. BETTERSTACK_API_TOKEN is required to run this " +
			"integration test against the REAL Better Stack API.")
	}
	t.Run(tt.name, func(t *testing.T) {
		settings := Settings{APIToken: os.Getenv("BETTERSTACK_API_TOKEN"), PageSize: 1}

		// create/update/delete actual check with REAL Better Stack API.
		m := New(settings)
		check, err := model.NewUptimeCheck("foo", tt.annotations)
		assert.NoError(t, err)
		if tt.wantDelete {
			if err := m.DeleteCheck(context.TODO(), *check); (err != nil) != tt.wantErr {
				t.Errorf("DeleteCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
			// give Better Stack some time to process the api call, just in case
			time.Sleep(5 * time.Second)

			existingCheckID, err := m.findCheck(*check)
			assert.NoError(t, err)
			assert.Equal(t, providers.CheckNotFound, existingCheckID)
		} else {
			if err := m.CreateOrUpdateCheck(context.TODO(), *check); (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdateCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
			// give Better Stack some time to process the api call, just in case
			time.Sleep(5 * time.Second)
			if tt.checkConfirmationPeriod {
				assertConfirmationPeriod(t, m, *check, tt.wantConfirmationPeriod)
			}
		}
	})
}

func assertConfirmationPeriod(t *testing.T, m *BetterStack, check model.UptimeCheck, expected int) {
	t.Helper()
	monitorID, err := m.findCheck(check)
	if !assert.NoError(t, err) || !assert.NotEqual(t, providers.CheckNotFound, monitorID) {
		return
	}
	monitor, err := m.client.getMonitor(monitorID)
	if !assert.NoError(t, err) || !assert.NotNil(t, monitor) {
		return
	}
	if !assert.NotNil(t, monitor.Data) {
		return
	}
	if !assert.NotNil(t, monitor.Data.Attributes) {
		return
	}
	assert.Equal(t, expected, monitor.Data.Attributes.ConfirmationPeriod)
}
