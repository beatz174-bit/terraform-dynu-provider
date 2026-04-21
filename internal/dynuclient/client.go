package dynuclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

type getRootResponse struct {
	apiResponse
	ID         int64  `json:"id"`
	Hostname   string `json:"hostname"`
	DomainName string `json:"domainName"`
	Node       string `json:"node"`
}

func (c *Client) ListDomains(ctx context.Context) ([]Domain, error) {
	var resp listDomainsResponse
	if err := c.doGET(ctx, "/dns", &resp); err != nil {
		return nil, err
	}
	return resp.Domains, nil
}

func (c *Client) GetDomainByID(ctx context.Context, domainID int64) (*Domain, error) {
	var resp getDomainResponse
	if err := c.doGET(ctx, fmt.Sprintf("/dns/%d", domainID), &resp); err != nil {
		return nil, err
	}
	return &resp.Domain, nil
}

func (c *Client) GetRootDomain(ctx context.Context, hostname string) (int64, string, error) {
	var resp getRootResponse
	if err := c.doGET(ctx, fmt.Sprintf("/dns/getroot/%s", url.PathEscape(hostname)), &resp); err != nil {
		return 0, "", err
	}

	if resp.ID == 0 || resp.DomainName == "" {
		return 0, "", errors.New("dynu API returned an incomplete root domain response")
	}

	return resp.ID, resp.DomainName, nil
}

func (c *Client) ListDNSRecords(ctx context.Context, domainID int64) ([]DNSRecord, error) {
	var resp listDNSRecordsResponse
	if err := c.doGET(ctx, fmt.Sprintf("/dns/%d/record", domainID), &resp); err != nil {
		return nil, err
	}
	return resp.DNSRecords, nil
}

func (c *Client) doGET(ctx context.Context, path string, target any) error {
	requestURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("API-Key", c.apiKey)

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

	if err := json.Unmarshal(payload, target); err != nil {
		return fmt.Errorf("failed to decode dynu API response: %w", err)
	}

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func parseAPIException(payload []byte) error {
	apiResult := apiResponse{}
	if err := json.Unmarshal(payload, &apiResult); err != nil || apiResult.Exception == nil {
		return nil
	}

	return fmt.Errorf("dynu API error %d (%s): %s", apiResult.Exception.StatusCode, apiResult.Exception.Type, apiResult.Exception.Message)
}
