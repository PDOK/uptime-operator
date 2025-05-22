package betterstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/PDOK/uptime-operator/internal/model"
)

//nolint:tagliatelle
type MetadataListRequest struct {
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`
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

func ListMetadata(client httpClient) (*MetadataListResponse, error) {
	url := fmt.Sprintf("%s/api/v3/metadata?owner_type=Monitor&per_page=%d", betterStackBaseURL, client.settings.PageSize)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Make initial HTTP request
	resp, err := client.call(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status %d, expected %d", resp.StatusCode, http.StatusOK)
	}

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

func (m MetadataListResponse) Next(client httpClient) (*MetadataListResponse, error) {
	if !m.HasNext() {
		return nil, nil
	}

	// Make HTTP request to the next URL
	resp, err := client.get(m.Pagination.Next)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status %d, expected %d", resp.StatusCode, http.StatusOK)
	}

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

func CreateOrUpdateMetadata(client httpClient, key string, monitorID int64, tags []string) error {
	metadataUpdateRequest := MetadataUpdateRequest{
		Key:       key,
		OwnerID:   strconv.FormatInt(monitorID, 10),
		OwnerType: "Monitor",
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
	resp, err := client.call(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK) {
		result, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got status %d, expected %d or %d. Body: %s", resp.StatusCode,
			http.StatusCreated, http.StatusOK, string(result))
	}
	return nil
}

func DeleteMetadata(client httpClient, key string, monitorID int64) error {
	metadataDeleteRequest := MetadataUpdateRequest{
		Key:       key,
		OwnerID:   strconv.FormatInt(monitorID, 10),
		OwnerType: "Monitor",
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
	resp, err := client.call(req)
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

//nolint:tagliatelle
type MonitorRequestHeader struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Value   string `json:"value"`
	Destroy bool   `json:"_destroy"`
}

//nolint:tagliatelle
type MonitorCreateRequest struct {
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

func CreateMonitor(client httpClient, check model.UptimeCheck) (int64, error) {
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
		return -1, err
	}
	req, err := http.NewRequest(http.MethodPost, betterStackBaseURL+"/api/v2/monitors", body)
	if err != nil {
		return -1, err
	}
	resp, err := client.call(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		result, _ := io.ReadAll(resp.Body)
		return -1, fmt.Errorf("got status %d, expected %d. Body: %s", resp.StatusCode, http.StatusCreated, string(result))
	}

	var monitorCreateResponse *MonitorCreateResponse
	err = json.NewDecoder(resp.Body).Decode(&monitorCreateResponse)
	if err != nil {
		return -1, err
	}
	monitorID, err := strconv.ParseInt(monitorCreateResponse.Data.ID, 10, 64)
	if err != nil {
		return -1, err
	}
	return monitorID, nil
}

func UpdateMonitor(client httpClient, check model.UptimeCheck, existingMonitor *MonitorGetResponse) error {
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
	err := json.NewEncoder(body).Encode(&updateRequest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/v2/monitors/%s", betterStackBaseURL, existingMonitor.Data.ID), body)
	if err != nil {
		return err
	}
	resp, err := client.call(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got status %d, expected %d", resp.StatusCode, http.StatusOK)
	}
	return nil
}

func DeleteMonitor(client httpClient, monitorID int64) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v2/monitors/%d", betterStackBaseURL, monitorID), nil)
	if err != nil {
		return err
	}
	resp, err := client.call(req)
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

func GetMonitor(client httpClient, monitorID int64) (*MonitorGetResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/monitors/%d", betterStackBaseURL, monitorID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.call(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status %d, expected %d", resp.StatusCode, http.StatusOK)
	}

	var existingMonitor *MonitorGetResponse
	err = json.NewDecoder(resp.Body).Decode(&existingMonitor)
	if err != nil {
		return nil, err
	}
	return existingMonitor, nil
}

//nolint:tagliatelle
type MonitorUpdateRequest struct {
	MonitorType       string                 `json:"monitor_type,omitempty"`
	URL               string                 `json:"url,omitempty"`
	PronounceableName string                 `json:"pronounceable_name,omitempty"`
	Port              int                    `json:"port,omitempty"`
	Email             bool                   `json:"email,omitempty"`
	Sms               bool                   `json:"sms,omitempty"`
	Call              bool                   `json:"call,omitempty"`
	RequiredKeyword   string                 `json:"required_keyword,omitempty"`
	CheckFrequency    int                    `json:"check_frequency,omitempty"`
	RequestHeaders    []MonitorRequestHeader `json:"request_headers,omitempty"`
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
