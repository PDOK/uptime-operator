package model

import "strings"

const (
	annotationBase              = "uptime.pdok.nl/"
	annotationId                = annotationBase + "id"
	annotationName              = annotationBase + "name"
	annotationUrl               = annotationBase + "url"
	annotationTags              = annotationBase + "tags"
	annotationRequestHeaders    = annotationBase + "request-headers"
	annotationStringContains    = annotationBase + "response-check-for-string-contains"
	annotationStringNotContains = annotationBase + "response-check-for-string-not-contains"
)

type UptimeCheck struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	URL               string            `json:"url"`
	Tags              []string          `json:"tags"`
	RequestHeaders    map[string]string `json:"request_headers"`
	StringContains    string            `json:"string_contains"`
	StringNotContains string            `json:"string_not_contains"`
}

func NewUptimeCheck(annotations map[string]string) *UptimeCheck {
	id, ok := annotations[annotationId]
	if !ok {
		return nil
	}
	return &UptimeCheck{
		ID:                id,
		Name:              annotations[annotationName],
		URL:               annotations[annotationUrl],
		Tags:              strings.Split(annotations[annotationTags], ","),
		RequestHeaders:    kvStringToMap(annotations[annotationRequestHeaders]),
		StringContains:    annotations[annotationStringContains],
		StringNotContains: annotations[annotationStringNotContains],
	}
}

func kvStringToMap(s string) map[string]string {
	kvPairs := strings.Split(s, ",")
	result := make(map[string]string)
	for _, kvPair := range kvPairs {
		parts := strings.Split(kvPair, ":")
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		result[key] = value
	}
	return result
}
