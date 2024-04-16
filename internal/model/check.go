package model

import (
	"slices"
	"strings"
)

const (
	AnnotationBase              = "uptime.pdok.nl"
	annotationID                = AnnotationBase + "/id"
	annotationName              = AnnotationBase + "/name"
	annotationURL               = AnnotationBase + "/url"
	annotationTags              = AnnotationBase + "/tags"
	annotationRequestHeaders    = AnnotationBase + "/request-headers"
	annotationStringContains    = AnnotationBase + "/response-check-for-string-contains"
	annotationStringNotContains = AnnotationBase + "/response-check-for-string-not-contains"

	// Indicate to humans that the given check is managed by the operator.
	tagManagedBy = "managed-by-uptime-operator"
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
	id, ok := annotations[annotationID]
	if !ok {
		return nil
	}
	check := &UptimeCheck{
		ID:                id,
		Name:              annotations[annotationName],
		URL:               annotations[annotationURL],
		Tags:              stringToSlice(annotations[annotationTags]),
		RequestHeaders:    kvStringToMap(annotations[annotationRequestHeaders]),
		StringContains:    annotations[annotationStringContains],
		StringNotContains: annotations[annotationStringNotContains],
	}
	if !slices.Contains(check.Tags, tagManagedBy) {
		check.Tags = append(check.Tags, tagManagedBy)
	}
	return check
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

func stringToSlice(s string) []string {
	var result []string
	splits := strings.Split(s, ",")
	for _, part := range splits {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}
