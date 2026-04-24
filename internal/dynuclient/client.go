package dynuclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.dynu.com/v2"

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type apiException struct {
	StatusCode int    `json:"statusCode"`
	Type       string `json:"type"`
	Message    string `json:"message"`
}

type APIError struct {
	StatusCode int
	Type       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("dynu API error %d (%s): %s", e.StatusCode, e.Type, e.Message)
}

type apiResponse struct {
	StatusCode int           `json:"statusCode"`
	Exception  *apiException `json:"exception"`
}

type Domain struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	UnicodeName       string `json:"unicodeName"`
	Token             string `json:"token"`
	State             string `json:"state"`
	Group             string `json:"group"`
	IPv4Address       string `json:"ipv4Address"`
	IPv6Address       string `json:"ipv6Address"`
	TTL               int64  `json:"ttl"`
	IPv4              bool   `json:"ipv4"`
	IPv6              bool   `json:"ipv6"`
	IPv4WildcardAlias bool   `json:"ipv4WildcardAlias"`
	IPv6WildcardAlias bool   `json:"ipv6WildcardAlias"`
	AllowZoneTransfer bool   `json:"allowZoneTransfer"`
	DNSSEC            bool   `json:"dnssec"`
	CreatedOn         string `json:"createdOn"`
	UpdatedOn         string `json:"updatedOn"`
}

type DNSRecord struct {
	ID         int64  `json:"id"`
	DomainID   int64  `json:"domainId"`
	DomainName string `json:"domainName"`
	NodeName   string `json:"nodeName"`
	Hostname   string `json:"hostname"`
	RecordType string `json:"recordType"`
	State      bool   `json:"state"`
	TTL        int64  `json:"ttl"`
	Content    string `json:"content"`
	UpdatedOn  string `json:"updatedOn"`
	Group      string `json:"group"`
	Host       string `json:"host"`
}

type CreateDNSRecordRequest struct {
	NodeName   string  `json:"nodeName,omitempty"`
	RecordType string  `json:"recordType"`
	Content    *string `json:"content,omitempty"`
	TTL        int64   `json:"ttl,omitempty"`
	State      *bool   `json:"state,omitempty"`
	Group      string  `json:"group,omitempty"`
	Host       string  `json:"host,omitempty"`
}

type UpdateDNSRecordRequest struct {
	NodeName   string  `json:"nodeName,omitempty"`
	RecordType string  `json:"recordType"`
	Content    *string `json:"content,omitempty"`
	TTL        int64   `json:"ttl,omitempty"`
	State      *bool   `json:"state,omitempty"`
	Group      string  `json:"group,omitempty"`
	Host       string  `json:"host,omitempty"`
}

type listDomainsResponse struct {
	apiResponse
	Domains []Domain `json:"domains"`
}

type getDomainResponse struct {
	apiResponse
	Domain
}

type listDNSRecordsResponse struct {
	apiResponse
	DNSRecords []DNSRecord `json:"dnsRecords"`
}

type getDNSRecordResponse struct {
	apiResponse
	DNSRecord
}

type getRootResponse struct {
	apiResponse
	ID         int64  `json:"id"`
	Hostname   string `json:"hostname"`
	DomainName string `json:"domainName"`
	Node       string `json:"node"`
}

func (c *Client) ListDomains(ctx context.Context) ([]Domain, error) {
	var resp listDomainsResponse
	if err := c.doRequest(ctx, http.MethodGet, "/dns", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Domains, nil
}

func (c *Client) GetDomainByID(ctx context.Context, domainID int64) (*Domain, error) {
	var resp getDomainResponse
	if err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/dns/%d", domainID), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Domain, nil
}

func (c *Client) GetRootDomain(ctx context.Context, hostname string) (int64, string, error) {
	var resp getRootResponse
	if err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/dns/getroot/%s", url.PathEscape(hostname)), nil, &resp); err != nil {
		return 0, "", err
	}

	if resp.ID == 0 || resp.DomainName == "" {
		return 0, "", errors.New("dynu API returned an incomplete root domain response")
	}

	return resp.ID, resp.DomainName, nil
}

func (c *Client) ListDNSRecords(ctx context.Context, domainID int64) ([]DNSRecord, error) {
	var resp listDNSRecordsResponse
	if err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/dns/%d/record", domainID), nil, &resp); err != nil {
		return nil, err
	}
	for i := range resp.DNSRecords {
		normalizeDNSRecord(&resp.DNSRecords[i])
	}
	return resp.DNSRecords, nil
}

func (c *Client) GetDNSRecord(ctx context.Context, domainID int64, recordID int64) (*DNSRecord, error) {
	var resp getDNSRecordResponse
	if err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/dns/%d/record/%d", domainID, recordID), nil, &resp); err != nil {
		return nil, err
	}
	normalizeDNSRecord(&resp.DNSRecord)
	return &resp.DNSRecord, nil
}

func (c *Client) CreateDNSRecord(ctx context.Context, domainID int64, req CreateDNSRecordRequest) (*DNSRecord, error) {
	var resp getDNSRecordResponse
	if err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/dns/%d/record", domainID), buildDNSRecordUpsertPayload(req.RecordType, req.NodeName, req.Content, req.TTL, req.State, req.Group, req.Host), &resp); err != nil {
		return nil, err
	}
	normalizeDNSRecord(&resp.DNSRecord)
	return &resp.DNSRecord, nil
}

func (c *Client) UpdateDNSRecord(ctx context.Context, domainID int64, recordID int64, req UpdateDNSRecordRequest) (*DNSRecord, error) {
	var resp getDNSRecordResponse
	if err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/dns/%d/record/%d", domainID, recordID), buildDNSRecordUpsertPayload(req.RecordType, req.NodeName, req.Content, req.TTL, req.State, req.Group, req.Host), &resp); err != nil {
		return nil, err
	}
	normalizeDNSRecord(&resp.DNSRecord)
	return &resp.DNSRecord, nil
}

func (c *Client) DeleteDNSRecord(ctx context.Context, domainID int64, recordID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/dns/%d/record/%d", domainID, recordID), nil, nil)
}

func (c *Client) doRequest(ctx context.Context, method string, path string, requestBody any, target any) error {
	requestURL := c.baseURL + path

	var bodyReader io.Reader
	if requestBody != nil {
		body, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to encode dynu API request: %w", err)
		}
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("API-Key", c.apiKey)
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	payload, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	apiErr := parseAPIException(payload)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		if apiErr != nil {
			return apiErr
		}
		return fmt.Errorf("dynu API returned status %d: %s", res.StatusCode, strings.TrimSpace(string(payload)))
	}

	if len(bytes.TrimSpace(payload)) > 0 && target != nil {
		if err := json.Unmarshal(payload, target); err != nil {
			return fmt.Errorf("failed to decode dynu API response: %w", err)
		}
	}

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func parseAPIException(payload []byte) error {
	apiResult := apiResponse{}
	if err := json.Unmarshal(payload, &apiResult); err != nil {
		return nil
	}

	if apiResult.Exception != nil {
		return &APIError{
			StatusCode: apiResult.Exception.StatusCode,
			Type:       apiResult.Exception.Type,
			Message:    apiResult.Exception.Message,
		}
	}

	// Some Dynu responses surface API failures at the top level instead of under exception.
	topLevel := apiException{}
	if err := json.Unmarshal(payload, &topLevel); err != nil || topLevel.StatusCode == 0 || topLevel.Type == "" {
		return nil
	}

	return &APIError{
		StatusCode: topLevel.StatusCode,
		Type:       topLevel.Type,
		Message:    topLevel.Message,
	}
}

type dnsRecordUpsertPayload struct {
	NodeName    string  `json:"nodeName,omitempty"`
	RecordType  string  `json:"recordType"`
	Content     *string `json:"content,omitempty"`
	IPv4Address string  `json:"ipv4Address,omitempty"`
	IPv6Address string  `json:"ipv6Address,omitempty"`
	TTL         int64   `json:"ttl,omitempty"`
	State       *bool   `json:"state,omitempty"`
	Group       string  `json:"group,omitempty"`
	Host        string  `json:"host,omitempty"`
}

func buildDNSRecordUpsertPayload(recordType string, nodeName string, content *string, ttl int64, state *bool, group string, host string) dnsRecordUpsertPayload {
	payload := dnsRecordUpsertPayload{
		NodeName:   nodeName,
		RecordType: recordType,
		Content:    content,
		TTL:        ttl,
		State:      state,
		Group:      group,
		Host:       host,
	}

	switch strings.ToUpper(strings.TrimSpace(recordType)) {
	case "A":
		if content != nil {
			payload.IPv4Address = *content
		}
	case "AAAA":
		if content != nil {
			payload.IPv6Address = *content
		}
	case "CNAME":
		if payload.Host == "" && content != nil {
			payload.Host = *content
		}
	}

	return payload
}

var zoneStyleContentPattern = regexp.MustCompile(`(?i)^\S+\.\s+\d+\s+IN\s+\S+\s+(.+)$`)

func normalizeDNSRecord(record *DNSRecord) {
	if record == nil {
		return
	}

	matches := zoneStyleContentPattern.FindStringSubmatch(strings.TrimSpace(record.Content))
	if len(matches) != 2 {
		return
	}

	record.Content = strings.TrimSpace(matches[1])
}
