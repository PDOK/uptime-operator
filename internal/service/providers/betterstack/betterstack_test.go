package betterstack

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/service/providers"
	"github.com/stretchr/testify/assert"
)

// Test against production better stack API. Please supply BETTERSTACK_API_TOKEN.
// This test creates one check, updates it and then deletes the check.
func TestAgainstREALBetterStackAPI(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		wantErr     bool
		wantDelete  bool
	}{
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
	tests := []struct {
		name        string
		annotations map[string]string
		wantErr     bool
		wantDelete  bool
	}{
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

func runIntegrationTest(t *testing.T, tt struct {
	name        string
	annotations map[string]string
	wantErr     bool
	wantDelete  bool
}) {
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
		}
	})
}

func TestToSupportedInterval(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "ZeroInput", input: 0, expected: 30},
		{name: "GreaterThanMaxSupported", input: 31, expected: 1800},
		{name: "MuchGreaterThanMaxSupported", input: 1000, expected: 1800},
		{name: "ExactMatch_60s", input: 1, expected: 60},
		{name: "ExactMatch_120s", input: 2, expected: 120},
		{name: "ExactMatch_180s", input: 3, expected: 180},
		{name: "ExactMatch_300s", input: 5, expected: 300},
		{name: "ExactMatch_600s", input: 10, expected: 600},
		{name: "ExactMatch_900s", input: 15, expected: 900},
		{name: "ExactMatch_1800s", input: 30, expected: 1800},
		{name: "Rounding_240s_roundsTo_180s", input: 4, expected: 180},
		{name: "Rounding_360s_roundsTo_300s", input: 6, expected: 300},
		{name: "Rounding_420s_roundsTo_300s", input: 7, expected: 300},
		{name: "Rounding_480s_roundsTo_600s", input: 8, expected: 600},
		{name: "Rounding_960s_roundsTo_900s", input: 16, expected: 900},
		{name: "Rounding_1320s_roundsTo_900s", input: 22, expected: 900},
		{name: "Rounding_1380s_roundsTo_1800s", input: 23, expected: 1800},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := toSupportedInterval(tt.input)
			if actual != tt.expected {
				t.Errorf("toSupportedInterval(%d) => expected %d, got %d", tt.input, tt.expected, actual)
			}
		})
	}
}
