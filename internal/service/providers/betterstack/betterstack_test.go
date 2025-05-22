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
