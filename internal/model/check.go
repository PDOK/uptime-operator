package model

import (
	"fmt"
	"slices"
	"strings"
)

const (
	OperatorName = "uptime-operator"

	// TagManagedBy Indicate to humans that the given check is managed by the operator.
	TagManagedBy = "managed-by-" + OperatorName

	AnnotationBase              = "uptime.pdok.nl"
	AnnotationFinalizer         = AnnotationBase + "/finalizer"
	AnnotationID                = AnnotationBase + "/id"
	AnnotationName              = AnnotationBase + "/name"
	AnnotationURL               = AnnotationBase + "/url"
	AnnotationTags              = AnnotationBase + "/tags"
	AnnotationRequestHeaders    = AnnotationBase + "/request-headers"
	AnnotationStringContains    = AnnotationBase + "/response-check-for-string-contains"
	AnnotationStringNotContains = AnnotationBase + "/response-check-for-string-not-contains"
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

func NewUptimeCheck(ingressName string, annotations map[string]string) (*UptimeCheck, error) {
	id, ok := annotations[AnnotationID]
	if !ok {
		return nil, fmt.Errorf("%s annotation not found on ingress route: %s", AnnotationID, ingressName)
	}
	name, ok := annotations[AnnotationName]
	if !ok {
		return nil, fmt.Errorf("%s annotation not found on ingress route: %s", AnnotationName, ingressName)
	}
	url, ok := annotations[AnnotationURL]
	if !ok {
		return nil, fmt.Errorf("%s annotation not found on ingress route %s", AnnotationURL, ingressName)
	}
	check := &UptimeCheck{
		ID:                id,
		Name:              name,
		URL:               url,
		Tags:              stringToSlice(annotations[AnnotationTags]),
		RequestHeaders:    kvStringToMap(annotations[AnnotationRequestHeaders]),
		StringContains:    annotations[AnnotationStringContains],
		StringNotContains: annotations[AnnotationStringNotContains],
	}
	if !slices.Contains(check.Tags, TagManagedBy) {
		check.Tags = append(check.Tags, TagManagedBy)
	}
	return check, nil
}

func kvStringToMap(s string) map[string]string {
	if s == "" {
		return nil
	}
	result := make(map[string]string)
	kvPairs := strings.Split(s, ",")
	for _, kvPair := range kvPairs {
		parts := strings.Split(kvPair, ":")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		result[key] = value
	}
	return result
}

func stringToSlice(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	splits := strings.Split(s, ",")
	for _, part := range splits {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}
