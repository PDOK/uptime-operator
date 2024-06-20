package providers

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/stretchr/testify/assert"
)

// Test against production pingdom API. Please supply PINGDOM_API_TOKEN + USER/INTEGRATION_ID.
// This test creates one check, updates it and then deletes the check.
func TestAgainstREALPingdomAPI(t *testing.T) {
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
				"uptime.pdok.nl/name":                                   "Test Check Richard",
				"uptime.pdok.nl/url":                                    "https://service.pdok.nl/cbs/landuse/wfs/v1_0?request=GetCapabilities&service=WFS",
				"uptime.pdok.nl/tags":                                   "tag1, tag2",
				"uptime.pdok.nl/request-headers":                        "key1:value1, key2:value2",
				"uptime.pdok.nl/response-check-for-string-contains":     "bla",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
		},
		{
			name: "Update check",
			annotations: map[string]string{
				"uptime.pdok.nl/id":                                     "3w2e9d804b2cd6bf18b8c0a6e1c04e46ac62b98c",
				"uptime.pdok.nl/name":                                   "Test Check Richard Updated",
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
				"uptime.pdok.nl/name":                                   "Test Check Richard Updated",
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
		if testing.Short() {
			t.Skip("skipping integration tests in short mode")
		}
		t.Run(tt.name, func(t *testing.T) {
			if os.Getenv("PINGDOM_API_TOKEN") == "" {
				assert.Fail(t, "PINGDOM_API_TOKEN is required")
			}
			settings := PingdomSettings{APIToken: os.Getenv("PINGDOM_API_TOKEN")}

			if os.Getenv("PINGDOM_USER_ID") != "" {
				userID, _ := strconv.Atoi(os.Getenv("PINGDOM_USER_ID"))
				settings.UserIDs = []int{userID}
			}
			if os.Getenv("PINGDOM_INTEGRATION_ID") != "" {
				integrationID, _ := strconv.Atoi(os.Getenv("PINGDOM_INTEGRATION_ID"))
				settings.IntegrationIDs = []int{integrationID}
			}

			// create/update/delete actual check with REAL pingdom API.
			m := NewPingdomUptimeProvider(settings)
			check, err := model.NewUptimeCheck("foo", tt.annotations)
			assert.NoError(t, err)
			if tt.wantDelete {
				if err := m.DeleteCheck(*check); (err != nil) != tt.wantErr {
					t.Errorf("DeleteCheck() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if err := m.CreateOrUpdateCheck(*check); (err != nil) != tt.wantErr {
					t.Errorf("CreateOrUpdateCheck() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			// give pingdom some time to process the api call, just in case
			time.Sleep(5 * time.Second)
		})
	}
}
