package model

import (
	"maps"
	"slices"
	"testing"
)

func TestNewUptimeCheckAnnotations(t *testing.T) {
	annotations := map[string]string{
		AnnotationID:                  "1234567890",
		AnnotationName:                "Test Check",
		AnnotationURL:                 "https://pdok.example",
		"example.com/provider-option": "unparsed value",
	}
	check, err := NewUptimeCheck("test-ingress", annotations)
	if err != nil {
		t.Fatalf("NewUptimeCheck() error = %v", err)
	}
	if !maps.Equal(check.Annotations, annotations) {
		t.Errorf("NewUptimeCheck().Annotations = %v, want %v", check.Annotations, annotations)
	}
	annotations["example.com/provider-option"] = "changed"
	if actual := check.Annotations["example.com/provider-option"]; actual != "unparsed value" {
		t.Errorf("NewUptimeCheck().Annotations should be a copy, got %q", actual)
	}
}

func TestNewUptimeCheck(t *testing.T) {
	tests := []struct {
		name        string
		ingressName string
		annotations map[string]string
		wantTags    []string
		wantErr     bool
	}{
		{
			name:        "All annotations present",
			ingressName: "test-ingress",
			annotations: map[string]string{
				"uptime.pdok.nl/id":                                     "1234567890",
				"uptime.pdok.nl/name":                                   "Test Check",
				"uptime.pdok.nl/url":                                    "https://pdok.example",
				"uptime.pdok.nl/tags":                                   "tag1, tag2",
				"uptime.pdok.nl/request-headers":                        "key1:value1, key2:value2",
				"uptime.pdok.nl/response-check-for-string-contains":     "test string",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
			wantTags: []string{"tag1", "tag2", TagManagedBy},
			wantErr:  false,
		},
		{
			name:        "Missing ID annotation",
			ingressName: "test-ingress",
			annotations: map[string]string{
				"uptime.pdok.nl/name":                                   "Test Check",
				"uptime.pdok.nl/url":                                    "https://pdok.example",
				"uptime.pdok.nl/tags":                                   "tag1, tag2",
				"uptime.pdok.nl/request-headers":                        "key1:value1, key2:value2",
				"uptime.pdok.nl/response-check-for-string-contains":     "test string",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
			wantTags: nil,
			wantErr:  true,
		},
		{
			name:        "Missing Name annotation",
			ingressName: "test-ingress",
			annotations: map[string]string{
				"uptime.pdok.nl/id":                                     "1234567890",
				"uptime.pdok.nl/url":                                    "https://pdok.example",
				"uptime.pdok.nl/tags":                                   "tag1, tag2",
				"uptime.pdok.nl/request-headers":                        "key1:value1, key2:value2",
				"uptime.pdok.nl/response-check-for-string-contains":     "test string",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
			wantTags: nil,
			wantErr:  true,
		},
		{
			name:        "Missing URL annotation",
			ingressName: "test-ingress",
			annotations: map[string]string{
				"uptime.pdok.nl/id":                                     "1234567890",
				"uptime.pdok.nl/name":                                   "Test Check",
				"uptime.pdok.nl/tags":                                   "tag1, tag2",
				"uptime.pdok.nl/request-headers":                        "key1:value1, key2:value2",
				"uptime.pdok.nl/response-check-for-string-contains":     "test string",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
			wantTags: nil,
			wantErr:  true,
		},
		{
			name:        "Missing tags annotation",
			ingressName: "test-ingress",
			annotations: map[string]string{
				"uptime.pdok.nl/id":                                     "1234567890",
				"uptime.pdok.nl/name":                                   "Test Check",
				"uptime.pdok.nl/url":                                    "https://pdok.example",
				"uptime.pdok.nl/request-headers":                        "key1:value1, key2:value2",
				"uptime.pdok.nl/response-check-for-string-contains":     "test string",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
			wantTags: []string{TagManagedBy},
			wantErr:  false,
		},
		{
			name:        "Missing request-headers annotation",
			ingressName: "test-ingress",
			annotations: map[string]string{
				"uptime.pdok.nl/id":   "1234567890",
				"uptime.pdok.nl/name": "Test Check",
				"uptime.pdok.nl/url":  "https://pdok.example",
				"uptime.pdok.nl/tags": "tag1, tag2",
				"uptime.pdok.nl/response-check-for-string-contains":     "test string",
				"uptime.pdok.nl/response-check-for-string-not-contains": "",
			},
			wantTags: []string{"tag1", "tag2", TagManagedBy},
			wantErr:  false,
		},
		{
			name:        "Duplicate tags",
			ingressName: "test-ingress",
			annotations: map[string]string{
				"uptime.pdok.nl/id":   "1234567890",
				"uptime.pdok.nl/name": "Test Check",
				"uptime.pdok.nl/url":  "https://pdok.example",
				"uptime.pdok.nl/tags": "tag1, tag2, tag1, tag3, tag2, , tag3",
			},
			wantTags: []string{"tag1", "tag2", "tag3", TagManagedBy},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := NewUptimeCheck(tt.ingressName, tt.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUptimeCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr || tt.wantTags == nil {
				return
			}
			if !slices.Equal(check.Tags, tt.wantTags) {
				t.Errorf("NewUptimeCheck().Tags = %v, want %v", check.Tags, tt.wantTags)
			}
		})
	}
}
