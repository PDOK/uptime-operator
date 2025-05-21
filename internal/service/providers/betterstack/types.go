package betterstack

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	req, err := http.NewRequest(http.MethodGet, betterStackBaseURL+"/api/v3/metadata?owner_type=Monitor", nil)
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
