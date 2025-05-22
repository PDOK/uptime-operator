package betterstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/PDOK/uptime-operator/internal/model"
	p "github.com/PDOK/uptime-operator/internal/service/providers"
)

const typeMonitor = "Monitor"

type Client struct {
	httpClient *http.Client
	settings   Settings
}

func (h Client) execRequest(req *http.Request, expectedStatus int) (*http.Response, error) {
	req.Header.Set(p.HeaderAuthorization, "Bearer "+h.settings.APIToken)
	req.Header.Set(p.HeaderAccept, p.MediaTypeJSON)
	req.Header.Set(p.HeaderContentType, p.MediaTypeJSON)
	req.Header.Add(p.HeaderUserAgent, model.OperatorName)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != expectedStatus {
		defer resp.Body.Close()
		var result []byte
		if resp.Body != nil {
			result, _ = io.ReadAll(resp.Body)
		}
		return nil, fmt.Errorf("got status %d, expected %d. Body: %s", resp.StatusCode, expectedStatus, string(result))
	}
	return resp, nil // caller should close resp.Body
}

func (h Client) execRequestIgnoreBody(req *http.Request, expectedStatus int) error {
	resp, err := h.execRequest(req, expectedStatus)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

//nolint:tagliatelle
type MetadataListResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes *struct {
			Key    string `json:"key"`
			Values []struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"values"`
			TeamName  string `json:"team_name"`
			OwnerID   string `json:"owner_id"`
			OwnerType string `json:"owner_type"`
		} `json:"attributes"`
	} `json:"data"`
	Pagination *struct {
		First string `json:"first"`
		Last  string `json:"last"`
		Prev  string `json:"prev"`
		Next  string `json:"next"`
	} `json:"pagination"`
}

// listMetadata https://betterstack.com/docs/uptime/api/list-all-existing-metadata/
func (h Client) listMetadata() (*MetadataListResponse, error) {
	url := fmt.Sprintf("%s/api/v3/metadata?owner_type=Monitor&per_page=%d", betterStackBaseURL, h.settings.PageSize)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Make initial HTTP request
	resp, err := h.execRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var metadata MetadataListResponse
	err = json.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (m MetadataListResponse) HasNext() bool {
	return m.Pagination != nil && m.Pagination.Next != ""
}

// Next paginate though metadata, see https://betterstack.com/docs/uptime/api/pagination/
func (m MetadataListResponse) Next(client Client) (*MetadataListResponse, error) {
	if !m.HasNext() {
		return nil, nil
	}

	// Make HTTP request to the next URL
	req, err := http.NewRequest(http.MethodGet, m.Pagination.Next, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.execRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var nextPage MetadataListResponse
	err = json.NewDecoder(resp.Body).Decode(&nextPage)
	if err != nil {
		return nil, err
	}
	return &nextPage, nil
}

type MetadataValue struct {
	Value string `json:"value"`
}

//nolint:tagliatelle
type MetadataUpdateRequest struct {
	Key       string          `json:"key"`
	Values    []MetadataValue `json:"values"`
	OwnerID   string          `json:"owner_id"`
	OwnerType string          `json:"owner_type"`
}

// createMetadata https://betterstack.com/docs/uptime/api/update-an-existing-metadata-record/
func (h Client) createMetadata(key string, monitorID int64, tags []string) error {
	metadataUpdateRequest := MetadataUpdateRequest{
		Key:       key,
		OwnerID:   strconv.FormatInt(monitorID, 10),
		OwnerType: typeMonitor,
	}
	for _, tag := range tags {
		metadataUpdateRequest.Values = append(metadataUpdateRequest.Values, MetadataValue{tag})
	}
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(&metadataUpdateRequest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, betterStackBaseURL+"/api/v3/metadata", body)
	if err != nil {
		return err
	}
	if err := h.execRequestIgnoreBody(req, http.StatusCreated); err != nil {
		return err
	}
	return nil
}

// updateMetadata https://betterstack.com/docs/uptime/api/update-an-existing-metadata-record/
func (h Client) updateMetadata(key string, monitorID int64, tags []string) error {
	metadataUpdateRequest := MetadataUpdateRequest{
		Key:       key,
		OwnerID:   strconv.FormatInt(monitorID, 10),
		OwnerType: typeMonitor,
	}
	for _, tag := range tags {
		metadataUpdateRequest.Values = append(metadataUpdateRequest.Values, MetadataValue{tag})
	}
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(&metadataUpdateRequest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, betterStackBaseURL+"/api/v3/metadata", body)
	if err != nil {
		return err
	}
	if err = h.execRequestIgnoreBody(req, http.StatusOK); err != nil {
		return err
	}
	return nil
}

// deleteMetadata https://betterstack.com/docs/uptime/api/update-an-existing-metadata-record/
func (h Client) deleteMetadata(key string, monitorID int64) error {
	metadataDeleteRequest := MetadataUpdateRequest{
		Key:       key,
		OwnerID:   strconv.FormatInt(monitorID, 10),
		OwnerType: typeMonitor,
		Values:    []MetadataValue{}, // empty values will result in delete of metadata record
	}
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(&metadataDeleteRequest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, betterStackBaseURL+"/api/v3/metadata", body)
	if err != nil {
		return err
	}
	if err = h.execRequestIgnoreBody(req, http.StatusNoContent); err != nil {
		return err
	}
	return nil
}

//nolint:tagliatelle
type MonitorRequestHeader struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Value   string `json:"value"`
	Destroy bool   `json:"_destroy"`
}

//nolint:tagliatelle
type MonitorCreateOrUpdateRequest struct {
	MonitorType       string                 `json:"monitor_type"`
	URL               string                 `json:"url"`
	PronounceableName string                 `json:"pronounceable_name"`
	Port              int                    `json:"port"`
	Email             bool                   `json:"email"`
	Sms               bool                   `json:"sms"`
	Call              bool                   `json:"call"`
	RequiredKeyword   string                 `json:"required_keyword"`
	CheckFrequency    int                    `json:"check_frequency"`
	RequestHeaders    []MonitorRequestHeader `json:"request_headers"`
}

type MonitorCreateResponse struct {
	Data struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"data"`
}

// createMonitor https://betterstack.com/docs/uptime/api/create-a-new-monitor/
func (h Client) createMonitor(check model.UptimeCheck) (int64, error) {
	createRequest := checkToMonitor(check)

	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(createRequest)
	if err != nil {
		return -1, err
	}
	req, err := http.NewRequest(http.MethodPost, betterStackBaseURL+"/api/v2/monitors", body)
	if err != nil {
		return -1, err
	}
	resp, err := h.execRequest(req, http.StatusCreated)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var createResponse *MonitorCreateResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	if err != nil {
		return -1, err
	}
	monitorID, err := strconv.ParseInt(createResponse.Data.ID, 10, 64)
	if err != nil {
		return -1, err
	}
	return monitorID, nil
}

// updateMonitor https://betterstack.com/docs/uptime/api/update-an-existing-monitor/
func (h Client) updateMonitor(check model.UptimeCheck, existingMonitor *MonitorGetResponse) error {
	updateRequest := checkToMonitor(check)

	if existingMonitor == nil || existingMonitor.Data == nil || existingMonitor.Data.Attributes == nil {
		return fmt.Errorf("invalid monitor response, expected values are nil: %v", existingMonitor)
	}
	// Remove all existing headers (since the API works with HTTP PATCH, to avoid duplicate headers).
	for _, existingHeader := range existingMonitor.Data.Attributes.RequestHeaders {
		updateRequest.RequestHeaders = append(updateRequest.RequestHeaders, MonitorRequestHeader{
			ID:      existingHeader.ID,
			Destroy: true,
		})
	}
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(&updateRequest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/v2/monitors/%s", betterStackBaseURL, existingMonitor.Data.ID), body)
	if err != nil {
		return err
	}
	if err = h.execRequestIgnoreBody(req, http.StatusOK); err != nil {
		return err
	}
	return nil
}

// deleteMonitor https://betterstack.com/docs/uptime/api/delete-an-existing-monitor/
func (h Client) deleteMonitor(monitorID int64) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v2/monitors/%d", betterStackBaseURL, monitorID), nil)
	if err != nil {
		return err
	}
	if err = h.execRequestIgnoreBody(req, http.StatusNoContent); err != nil {
		return err
	}
	return nil
}

//nolint:tagliatelle
type MonitorGetResponse struct {
	Data *struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes *struct {
			URL               string                 `json:"url"`
			PronounceableName string                 `json:"pronounceable_name"`
			MonitorType       string                 `json:"monitor_type"`
			RequiredKeyword   string                 `json:"required_keyword"`
			CheckFrequency    int                    `json:"check_frequency"`
			RequestHeaders    []MonitorRequestHeader `json:"request_headers"`
		} `json:"attributes"`
	} `json:"data"`
}

func (h Client) GetMonitor(monitorID int64) (*MonitorGetResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/monitors/%d", betterStackBaseURL, monitorID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.execRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var existingMonitor *MonitorGetResponse
	err = json.NewDecoder(resp.Body).Decode(&existingMonitor)
	if err != nil {
		return nil, err
	}
	return existingMonitor, nil
}

func checkToMonitor(check model.UptimeCheck) MonitorCreateOrUpdateRequest {
	var request MonitorCreateOrUpdateRequest
	switch {
	case check.StringContains != "":
		request = MonitorCreateOrUpdateRequest{
			MonitorType:     "keyword",
			RequiredKeyword: check.StringContains,
		}
	case check.StringNotContains != "":
		request = MonitorCreateOrUpdateRequest{
			MonitorType:     "keyword_absence",
			RequiredKeyword: check.StringNotContains,
		}
	default:
		request = MonitorCreateOrUpdateRequest{
			MonitorType: "status",
		}
	}
	request.URL = check.URL
	request.PronounceableName = check.Name
	request.Port = 443
	request.CheckFrequency = toSupportedInterval(check.Interval)
	request.Email = false
	request.Sms = false
	request.Call = false
	for name, value := range check.RequestHeaders {
		request.RequestHeaders = append(request.RequestHeaders, MonitorRequestHeader{
			Name:  name,
			Value: value,
		})
	}
	return request
}
