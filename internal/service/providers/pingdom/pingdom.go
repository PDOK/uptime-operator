package pingdom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	classiclog "log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/service/providers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const pingdomURL = "https://api.pingdom.com/api/3.1/checks"
const customIDPrefix = "id:"

const headerReqLimitShort = "Req-Limit-Short"
const headerReqLimitLong = "Req-Limit-Long"

type Settings struct {
	APIToken       string
	UserIDs        []int
	IntegrationIDs []int
}

type Pingdom struct {
	settings   Settings
	httpClient *http.Client
}

// New creates a Pingdom
func New(settings Settings) *Pingdom {
	if settings.APIToken == "" {
		classiclog.Fatal("Pingdom API token is not provided")
	}
	return &Pingdom{
		settings:   settings,
		httpClient: &http.Client{Timeout: time.Duration(5) * time.Minute},
	}
}

// CreateOrUpdateCheck create the given check with Pingdom, or update an existing check. Needs to be idempotent!
func (p *Pingdom) CreateOrUpdateCheck(ctx context.Context, check model.UptimeCheck) (err error) {
	existingCheckID, err := p.findCheck(ctx, check)
	if err != nil {
		return err
	}
	if existingCheckID == providers.CheckNotFound {
		err = p.createCheck(ctx, check)
	} else {
		err = p.updateCheck(ctx, existingCheckID, check)
	}
	return err
}

// DeleteCheck deletes the given check from Pingdom
func (p *Pingdom) DeleteCheck(ctx context.Context, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("deleting check", "check", check)

	existingCheckID, err := p.findCheck(ctx, check)
	if err != nil {
		return err
	}
	if existingCheckID == providers.CheckNotFound {
		log.FromContext(ctx).Info(fmt.Sprintf("check with ID '%s' is already deleted", check.ID))
		return nil
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", pingdomURL, existingCheckID), nil)
	if err != nil {
		return err
	}
	resp, err := p.execRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		resultBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got status %d, expected HTTP OK when deleting existing check. Error %s", resp.StatusCode, resultBody)
	}
	return nil
}

func (p *Pingdom) findCheck(ctx context.Context, check model.UptimeCheck) (int64, error) {
	result := providers.CheckNotFound

	// list all checks managed by uptime-operator. Can be at most 25.000, which is probably sufficient.
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?include_tags=true&limit=25000&tags=%s", pingdomURL, model.TagManagedBy), nil)
	if err != nil {
		return result, err
	}
	req.Header.Add(providers.HeaderAccept, providers.MediaTypeJSON)
	resp, err := p.execRequest(ctx, req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("got status %d, expected HTTP OK when listing existing checks", resp.StatusCode)
	}

	checksResponse := make(map[string]any)
	err = json.NewDecoder(resp.Body).Decode(&checksResponse)
	if err != nil {
		return result, err
	}

	pingdomChecks := checksResponse["checks"].([]any)
	for _, rawCheck := range pingdomChecks {
		pingdomCheck := rawCheck.(map[string]any)
		tags := pingdomCheck["tags"]
		if tags == nil {
			continue
		}
		for _, rawTag := range tags.([]any) {
			tag := rawTag.(map[string]any)
			if tag == nil {
				continue
			}
			tagName := tag["name"].(string)
			if strings.HasSuffix(tagName, check.ID) {
				// bingo, we've found the Pingdom check based on our custom ID (check.ID which is stored in a Pingdom tag).
				// now we return the actual Pingdom ID which we need for updates/deletes/etc.
				pingdomCheckID := pingdomCheck["id"]
				if pingdomCheckID != nil && pingdomCheckID.(float64) > 0 {
					result = int64(pingdomCheckID.(float64))
				}
			}
		}
	}
	return result, nil
}

func (p *Pingdom) createCheck(ctx context.Context, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("creating check", "check", check)

	message, err := p.checkToJSON(check, true)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, pingdomURL, bytes.NewBuffer(message))
	if err != nil {
		return err
	}
	err = p.execRequestWithBody(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pingdom) updateCheck(ctx context.Context, existingPingdomID int64, check model.UptimeCheck) error {
	log.FromContext(ctx).Info("updating check", "check", check, "pingdom ID", existingPingdomID)

	message, err := p.checkToJSON(check, false)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", pingdomURL, existingPingdomID), bytes.NewBuffer(message))
	if err != nil {
		return err
	}
	err = p.execRequestWithBody(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pingdom) checkToJSON(check model.UptimeCheck, includeType bool) ([]byte, error) {
	checkURL, err := url.ParseRequestURI(check.URL)
	if err != nil {
		return nil, err
	}
	port, err := getPort(checkURL)
	if err != nil {
		return nil, err
	}
	relativeURL := checkURL.Path
	if checkURL.RawQuery != "" {
		relativeURL += "?" + checkURL.RawQuery
	}

	// add the check id (from the k8s annotation) as a tag, so
	// we can latter retrieve the check during update or delete.
	check.Tags = append(check.Tags, customIDPrefix+check.ID)

	// tags can be at most 64 chars long, cut off longer ones
	for k := range check.Tags {
		tag := check.Tags[k]
		if len(tag) > 64 {
			tag = tag[:64]
		}
		check.Tags[k] = tag
	}

	message := map[string]any{
		"name":       check.Name,
		"host":       checkURL.Hostname(),
		"url":        relativeURL,
		"encryption": true, // assume all checks run over HTTPS
		"port":       port,
		"resolution": check.Interval,
		"tags":       check.Tags,
	}
	if includeType {
		// update messages shouldn't include 'type', since the type of check can't be modified in Pingdom.
		message["type"] = "http"
	}
	if len(p.settings.UserIDs) > 0 {
		message["userids"] = p.settings.UserIDs
	}
	if len(p.settings.IntegrationIDs) > 0 {
		message["integrationids"] = p.settings.IntegrationIDs
	}

	// request header need to be submitted in numbered JSON keys
	// for example "requestheader1": key:value, "requestheader2": key:value, etc
	var headers []string
	for header := range check.RequestHeaders {
		headers = append(headers, header)
	}
	sort.Strings(headers)
	for i, header := range headers {
		message[fmt.Sprintf("requestheader%d", i)] = fmt.Sprintf("%s:%s", header, check.RequestHeaders[header])
	}

	// Pingdom doesn't allow both "shouldcontain" and "shouldnotcontain"
	if check.StringContains != "" {
		message["shouldcontain"] = check.StringContains
	} else if check.StringNotContains != "" {
		message["shouldnotcontain"] = check.StringNotContains
	}

	return json.Marshal(message)
}

func (p *Pingdom) execRequestWithBody(ctx context.Context, req *http.Request) error {
	req.Header.Add(providers.HeaderContentType, providers.MediaTypeJSON)
	resp, err := p.execRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		resultBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got http status %d, while expected 200. Error: %s", resp.StatusCode, resultBody)
	}
	return nil
}

func (p *Pingdom) execRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Add(providers.HeaderAuthorization, "Bearer "+p.settings.APIToken)
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	rateLimitErr := errors.Join(
		handleRateLimits(ctx, resp.Header.Get(headerReqLimitShort)),
		handleRateLimits(ctx, resp.Header.Get(headerReqLimitLong)),
	)
	return resp, rateLimitErr
}

func handleRateLimits(ctx context.Context, rateLimitHeader string) error {
	remaining, resetTime, err := parseRateLimitHeader(rateLimitHeader)
	if err != nil {
		return err
	}
	if remaining < 25 {
		log.FromContext(ctx).Info(
			fmt.Sprintf("Waiting for %d seconds to avoid hitting Pingdom rate limit", resetTime+1),
			rateLimitHeader, remaining)

		time.Sleep(time.Duration(remaining+1) * time.Second)
	}
	return nil
}

func parseRateLimitHeader(header string) (remaining int, resetTime int, err error) {
	_, err = fmt.Sscanf(header, "Remaining: %d Time until reset: %d", &remaining, &resetTime)
	return
}

func getPort(checkURL *url.URL) (int, error) {
	port := checkURL.Port()
	if port == "" {
		port = "443"
	}
	return strconv.Atoi(port)
}
