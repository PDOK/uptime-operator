package model

import (
	"fmt"
	"slices"
	"strings"
)

const (
	OperatorName = "uptime-operator"

	// Indicate to humans that the given check is managed by the operator.
	tagManagedBy = "managed-by-" + OperatorName

	AnnotationBase              = "uptime.pdok.nl"
	AnnotationFinalizer         = AnnotationBase + "/finalizer"
	annotationID                = AnnotationBase + "/id"
	annotationName              = AnnotationBase + "/name"
	annotationURL               = AnnotationBase + "/url"
	annotationTags              = AnnotationBase + "/tags"
	annotationRequestHeaders    = AnnotationBase + "/request-headers"
	annotationStringContains    = AnnotationBase + "/response-check-for-string-contains"
	annotationStringNotContains = AnnotationBase + "/response-check-for-string-not-contains"
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
	id, ok := annotations[annotationID]
	if !ok {
		return nil, fmt.Errorf("%s annotation not found on ingress route: %s", annotationID, ingressName)
	}
	name, ok := annotations[annotationName]
	if !ok {
		return nil, fmt.Errorf("%s annotation not found on ingress route: %s", annotationName, ingressName)
	}
	url, ok := annotations[annotationURL]
	if !ok {
		return nil, fmt.Errorf("%s annotation not found on ingress route %s", annotationURL, ingressName)
	}
	check := &UptimeCheck{
		ID:                id,
		Name:              name,
		URL:               url,
		Tags:              stringToSlice(annotations[annotationTags]),
		RequestHeaders:    kvStringToMap(annotations[annotationRequestHeaders]),
		StringContains:    annotations[annotationStringContains],
		StringNotContains: annotations[annotationStringNotContains],
	}
	if !slices.Contains(check.Tags, tagManagedBy) {
		check.Tags = append(check.Tags, tagManagedBy)
	}
	return check, nil
}

func kvStringToMap(s string) map[string]string {
	kvPairs := strings.Split(s, ",")
	result := make(map[string]string)
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
	var result []string
	splits := strings.Split(s, ",")
	for _, part := range splits {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}
