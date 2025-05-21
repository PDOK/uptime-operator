package betterstack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	classiclog "log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
	p "github.com/PDOK/uptime-operator/internal/service/providers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const betterStackBaseURL = "https://uptime.betterstack.com"

type httpClient struct {
	client   *http.Client
	settings Settings
}

func (h httpClient) call(req *http.Request) (*http.Response, error) {
	req.Header.Set(p.HeaderAuthorization, "Bearer "+h.settings.APIToken)
	req.Header.Set(p.HeaderAccept, p.MediaTypeJSON)
	req.Header.Set(p.HeaderContentType, p.MediaTypeJSON)
	return h.client.Do(req)
}

func (h httpClient) get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(p.HeaderAuthorization, "Bearer "+h.settings.APIToken)
	req.Header.Set(p.HeaderAccept, p.MediaTypeJSON)
	req.Header.Set(p.HeaderContentType, p.MediaTypeJSON)
	return h.client.Do(req)
}

type Settings struct {
	APIToken string
	PageSize int
}

type BetterStack struct {
	httpClient httpClient
}

// New creates a BetterStack
func New(settings Settings) *BetterStack {
	if settings.APIToken == "" {
		classiclog.Fatal("Better Stack API token is not provided")
	}
	if settings.PageSize < 1 {
		settings.PageSize = 50 // def
	}
	return &BetterStack{
		httpClient: httpClient{
			client:   &http.Client{Timeout: time.Duration(5) * time.Minute},
			settings: settings,
		},
	}
}

// CreateOrUpdateCheck create the given check with Better Stack, or update an existing check. Needs to be idempotent!
func (b *BetterStack) CreateOrUpdateCheck(ctx context.Context, check model.UptimeCheck) (err error) {
	existingCheckID, err := b.findCheck(check)
	if err != nil {
		return err
	}
	if existingCheckID == p.CheckNotFound {
		err = b.createCheck(ctx, check)
	} else {
		err = b.updateCheck(ctx, existingCheckID, check)
	}
	return err
}

// DeleteCheck deletes the given check from Better Stack
func (b *BetterStack) DeleteCheck(ctx context.Context, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("deleting check", "check", check)

	existingCheckID, err := b.findCheck(check)
	if err != nil {
		return err
	}
	if existingCheckID == p.CheckNotFound {
		log.FromContext(ctx).Info(fmt.Sprintf("check with ID '%s' is already deleted", check.ID))
		return nil
	}

	metadataDeleteRequest := MetadataUpdateRequest{
		Key:       check.ID,
		OwnerID:   strconv.FormatInt(existingCheckID, 10),
		OwnerType: "Monitor",
		Values:    []MetadataValue{}, // empty values will result in delete of metadata record
	}
	body := &bytes.Buffer{}
	err = json.NewEncoder(body).Encode(&metadataDeleteRequest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, betterStackBaseURL+"/api/v3/metadata", body)
	if err != nil {
		return err
	}
	resp, err := b.httpClient.call(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		result, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got status %d, expected %d. Body: %s", resp.StatusCode, http.StatusNoContent, string(result))
	}

	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v2/monitors/%d", betterStackBaseURL, existingCheckID), nil)
	if err != nil {
		return err
	}
	resp, err = b.httpClient.call(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		result, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got status %d, expected %d. Body: %s", resp.StatusCode, http.StatusNoContent, string(result))
	}

	return nil
}

func (b *BetterStack) findCheck(check model.UptimeCheck) (int64, error) {
	result := p.CheckNotFound
	metadata, err := ListMetadata(b.httpClient)
	if err != nil {
		return result, err
	}
	for {
		for _, md := range metadata.Data {
			if md.Attributes != nil && md.Attributes.Key == check.ID {
				result, err = strconv.ParseInt(md.Attributes.OwnerID, 10, 64)
				if err != nil {
					return result, fmt.Errorf("failed to parse monitor ID %s to integer", md.Attributes.OwnerID)
				}
				return result, nil
			}
		}
		if !metadata.HasNext() {
			break // exit infinite loop
		}
		metadata, err = metadata.Next(b.httpClient)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

//nolint:funlen // TODO remove after refactor
func (b *BetterStack) createCheck(ctx context.Context, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("creating check", "check", check)

	// https://betterstack.com/docs/uptime/api/create-a-new-monitor/
	var monitorCreateRequest MonitorCreateRequest
	switch {
	case check.StringContains != "":
		monitorCreateRequest = MonitorCreateRequest{
			MonitorType:     "keyword",
			RequiredKeyword: check.StringContains,
		}
	case check.StringNotContains != "":
		monitorCreateRequest = MonitorCreateRequest{
			MonitorType:     "keyword_absence",
			RequiredKeyword: check.StringNotContains,
		}
	default:
		monitorCreateRequest = MonitorCreateRequest{
			MonitorType: "status",
		}
	}
	monitorCreateRequest.URL = check.URL
	monitorCreateRequest.PronounceableName = check.Name
	monitorCreateRequest.Port = 443
	monitorCreateRequest.CheckFrequency = toSupportedInterval(check.Interval)
	monitorCreateRequest.Email = false
	monitorCreateRequest.Sms = false
	monitorCreateRequest.Call = false
	for name, value := range check.RequestHeaders {
		monitorCreateRequest.RequestHeaders = append(monitorCreateRequest.RequestHeaders, MonitorRequestHeader{
			Name:  name,
			Value: value,
		})
	}

	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(monitorCreateRequest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, betterStackBaseURL+"/api/v2/monitors", body)
	if err != nil {
		return err
	}
	resp, err := b.httpClient.call(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		result, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got status %d, expected %d. Body: %s", resp.StatusCode, http.StatusCreated, string(result))
	}

	var monitorCreateResponse *MonitorCreateResponse
	err = json.NewDecoder(resp.Body).Decode(&monitorCreateResponse)
	if err != nil {
		return err
	}

	metadataUpdateRequest := MetadataUpdateRequest{
		Key:       check.ID,
		OwnerID:   monitorCreateResponse.Data.ID,
		OwnerType: "Monitor",
	}
	for _, tag := range check.Tags {
		metadataUpdateRequest.Values = append(metadataUpdateRequest.Values, MetadataValue{tag})
	}
	body = &bytes.Buffer{}
	err = json.NewEncoder(body).Encode(&metadataUpdateRequest)
	if err != nil {
		return err
	}
	req, err = http.NewRequest(http.MethodPost, betterStackBaseURL+"/api/v3/metadata", body)
	if err != nil {
		return err
	}
	resp, err = b.httpClient.call(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		result, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got status %d, expected %d. Body: %s", resp.StatusCode, http.StatusCreated, string(result))
	}

	return nil
}

//nolint:cyclop,funlen // TODO fix after refactor
func (b *BetterStack) updateCheck(ctx context.Context, existingCheckID int64, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("updating check", "check", check, "betterstack ID", existingCheckID)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/monitors/%d", betterStackBaseURL, existingCheckID), nil)
	if err != nil {
		return err
	}

	resp, err := b.httpClient.call(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got status %d, expected %d", resp.StatusCode, http.StatusOK)
	}

	// Parse response
	var existingMonitor MonitorGetResponse
	err = json.NewDecoder(resp.Body).Decode(&existingMonitor)
	if err != nil {
		return err
	}

	// Update monitor
	var updateRequest MonitorUpdateRequest
	switch {
	case check.StringContains != "":
		updateRequest = MonitorUpdateRequest{
			MonitorType:     "keyword",
			RequiredKeyword: check.StringContains,
		}
	case check.StringNotContains != "":
		updateRequest = MonitorUpdateRequest{
			MonitorType:     "keyword_absence",
			RequiredKeyword: check.StringNotContains,
		}
	default:
		updateRequest = MonitorUpdateRequest{
			MonitorType: "status",
		}
	}
	updateRequest.URL = check.URL
	updateRequest.PronounceableName = check.Name
	updateRequest.Port = 443
	updateRequest.CheckFrequency = toSupportedInterval(check.Interval)
	updateRequest.Email = false
	updateRequest.Sms = false
	updateRequest.Call = false
	// Add (new) headers
	for name, value := range check.RequestHeaders {
		updateRequest.RequestHeaders = append(updateRequest.RequestHeaders, MonitorRequestHeader{
			Name:  name,
			Value: value,
		})
	}
	// Remove all existing headers
	for _, existingHeader := range existingMonitor.Data.Attributes.RequestHeaders {
		updateRequest.RequestHeaders = append(updateRequest.RequestHeaders, MonitorRequestHeader{
			ID:      existingHeader.ID,
			Destroy: true,
		})
	}
	body := &bytes.Buffer{}
	err = json.NewEncoder(body).Encode(&updateRequest)
	if err != nil {
		return err
	}
	req, err = http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/v2/monitors/%d", betterStackBaseURL, existingCheckID), body)
	if err != nil {
		return err
	}
	resp, err = b.httpClient.call(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got status %d, expected %d", resp.StatusCode, http.StatusOK)
	}

	// Update tags metadata
	metadataUpdateRequest := MetadataUpdateRequest{
		Key:       check.ID,
		OwnerID:   strconv.FormatInt(existingCheckID, 10),
		OwnerType: "Monitor",
	}
	for _, tag := range check.Tags {
		metadataUpdateRequest.Values = append(metadataUpdateRequest.Values, MetadataValue{tag})
	}
	body = &bytes.Buffer{}
	err = json.NewEncoder(body).Encode(&metadataUpdateRequest)
	if err != nil {
		return err
	}
	req, err = http.NewRequest(http.MethodPost, betterStackBaseURL+"/api/v3/metadata", body)
	if err != nil {
		return err
	}
	resp, err = b.httpClient.call(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		result, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got status %d, expected %d. Body: %s", resp.StatusCode, http.StatusCreated, string(result))
	}
	return nil
}

func toSupportedInterval(intervalInMin int) int {
	// Better Stack only accepts a specific sets of intervals
	supportedIntervals := []int{30, 45, 60, 120, 180, 300, 600, 900, 1800}

	intervalInSec := intervalInMin * 60
	if intervalInSec <= 0 {
		return supportedIntervals[0] // use the smallest supported interval
	}
	if intervalInSec > supportedIntervals[len(supportedIntervals)-1] {
		return supportedIntervals[len(supportedIntervals)-1] // use the largest supported interval
	}

	nearestInterval := supportedIntervals[0]
	prevDiff := math.MaxInt

	// use nearest supported interval
	for _, si := range supportedIntervals {
		diff := int(math.Abs(float64(intervalInSec - si)))
		if diff < prevDiff {
			prevDiff = diff
			nearestInterval = si
		}
	}
	return nearestInterval
}
