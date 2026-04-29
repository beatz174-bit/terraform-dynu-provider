package dynuclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dynu/terraform-provider-dynu/internal/dynuclient"
	"github.com/dynu/terraform-provider-dynu/internal/testutil/fakedynu"
)

func TestClientListDomainsSuccess(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	domains, err := client.ListDomains(context.Background())
	if err != nil {
		t.Fatalf("ListDomains() error = %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(domains))
	}
}

func TestClientDoRequestNon2xxStatus(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()
	fake.SetRawResponse("/dns", 401, `{"message":"nope"}`)

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	_, err := client.ListDomains(context.Background())
	if err == nil || !strings.Contains(err.Error(), "status 401") {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestClientDoRequestAPIExceptionPayload(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()
	fake.SetAPIError("/dns", fakedynu.APIError{HTTPStatus: 400, StatusCode: 400, Type: "Validation Exception", Message: "bad hostname"})

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	_, err := client.ListDomains(context.Background())
	if err == nil || !strings.Contains(err.Error(), "Validation Exception") {
		t.Fatalf("expected API exception error, got %v", err)
	}
}

func TestClientDoRequestMalformedJSON(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()
	fake.SetRawResponse("/dns", 200, `{"statusCode":200,"domains":[`)

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	_, err := client.ListDomains(context.Background())
	if err == nil || !strings.Contains(err.Error(), "failed to decode dynu API response") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestClientGetRootDomainIncompleteResponse(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()
	fake.SetRawResponse("/dns/getroot/www.a.example.com", 200, `{"statusCode":200,"id":0,"domainName":""}`)

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	_, _, err := client.GetRootDomain(context.Background(), "www.a.example.com")
	if err == nil || !strings.Contains(err.Error(), "incomplete root domain response") {
		t.Fatalf("expected incomplete response error, got %v", err)
	}
}

func TestClientGetRootDomainEscapesHostname(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()
	fake.SetAPIError("/dns/getroot/spaces%20in%20name.example.com", fakedynu.APIError{HTTPStatus: 404, StatusCode: 404, Type: "Not Found", Message: "hostname not found"})

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	_, _, err := client.GetRootDomain(context.Background(), "spaces in name.example.com")
	if err == nil || !strings.Contains(err.Error(), "hostname not found") {
		t.Fatalf("expected hostname not found error, got %v", err)
	}
}

func TestClientDNSRecordCRUD(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	state := true
	created, err := client.CreateDNSRecord(context.Background(), 1001, dynuclient.CreateDNSRecordRequest{
		NodeName:   "api",
		RecordType: "TXT",
		Content:    stringPointer("created"),
		TTL:        120,
		State:      &state,
		Group:      "integration",
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord() error = %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected created record id")
	}

	got, err := client.GetDNSRecord(context.Background(), 1001, created.ID)
	if err != nil {
		t.Fatalf("GetDNSRecord() error = %v", err)
	}
	if got.Content != "created" {
		t.Fatalf("expected content created, got %q", got.Content)
	}

	updated, err := client.UpdateDNSRecord(context.Background(), 1001, created.ID, dynuclient.UpdateDNSRecordRequest{
		NodeName:   "api",
		RecordType: "TXT",
		Content:    stringPointer("updated"),
		TTL:        180,
		State:      &state,
	})
	if err != nil {
		t.Fatalf("UpdateDNSRecord() error = %v", err)
	}
	if updated.Content != "updated" || updated.TTL != 180 {
		t.Fatalf("unexpected update response: %#v", updated)
	}

	if err := client.DeleteDNSRecord(context.Background(), 1001, created.ID); err != nil {
		t.Fatalf("DeleteDNSRecord() error = %v", err)
	}

	_, err = client.GetDNSRecord(context.Background(), 1001, created.ID)
	if err == nil || !strings.Contains(err.Error(), "record not found") {
		t.Fatalf("expected record not found after delete, got %v", err)
	}
}

func TestClientDNSRecordWriteAPIError(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()
	fake.SetAPIError("/dns/1001/record", fakedynu.APIError{HTTPStatus: 400, StatusCode: 400, Type: "Validation Exception", Message: "recordType invalid"})

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	_, err := client.CreateDNSRecord(context.Background(), 1001, dynuclient.CreateDNSRecordRequest{RecordType: "", Content: stringPointer("x")})
	if err == nil || !strings.Contains(err.Error(), "Validation Exception") {
		t.Fatalf("expected validation API error, got %v", err)
	}
}

func TestClientCreateDNSRecordSendsIPv4AddressForARecord(t *testing.T) {
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/dns/1001/record" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"statusCode": 200,
			"id":         10,
			"domainId":   1001,
			"domainName": "example.com",
			"nodeName":   "www",
			"hostname":   "www.example.com",
			"recordType": "A",
			"content":    "167.179.167.166",
			"ttl":        300,
			"state":      true,
		})
	}))
	defer server.Close()

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(server.URL), dynuclient.WithHTTPClient(server.Client()))
	_, err := client.CreateDNSRecord(context.Background(), 1001, dynuclient.CreateDNSRecordRequest{
		NodeName:   "www",
		RecordType: "A",
		Content:    stringPointer("167.179.167.166"),
		TTL:        300,
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord() error = %v", err)
	}

	if captured["ipv4Address"] != "167.179.167.166" {
		t.Fatalf("expected ipv4Address in payload, got %#v", captured)
	}
	if _, ok := captured["content"]; ok {
		t.Fatalf("expected content to be omitted for A record, got %#v", captured["content"])
	}
	if _, ok := captured["ipv6Address"]; ok {
		t.Fatalf("expected ipv6Address to be omitted for A record, got %#v", captured["ipv6Address"])
	}
}

func TestClientUpdateDNSRecordSendsIPv6AddressForAAAARecord(t *testing.T) {
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/dns/1001/record/2002" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"statusCode": 200,
			"id":         2002,
			"domainId":   1001,
			"domainName": "example.com",
			"nodeName":   "www",
			"hostname":   "www.example.com",
			"recordType": "AAAA",
			"content":    "2001:db8::123",
			"ttl":        90,
			"state":      true,
		})
	}))
	defer server.Close()

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(server.URL), dynuclient.WithHTTPClient(server.Client()))
	_, err := client.UpdateDNSRecord(context.Background(), 1001, 2002, dynuclient.UpdateDNSRecordRequest{
		NodeName:   "www",
		RecordType: "AAAA",
		Content:    stringPointer("2001:db8::123"),
		TTL:        90,
	})
	if err != nil {
		t.Fatalf("UpdateDNSRecord() error = %v", err)
	}

	if captured["recordType"] != "AAAA" {
		t.Fatalf("expected recordType AAAA, got %#v", captured["recordType"])
	}
	if captured["ipv6Address"] != "2001:db8::123" {
		t.Fatalf("expected ipv6Address in payload, got %#v", captured)
	}
	if _, ok := captured["content"]; ok {
		t.Fatalf("expected content to be omitted for AAAA record, got %#v", captured["content"])
	}
	if _, ok := captured["ipv4Address"]; ok {
		t.Fatalf("expected ipv4Address to be omitted for AAAA record, got %#v", captured["ipv4Address"])
	}
}

func TestClientCreateDNSRecordSendsHostForCNAMERecord(t *testing.T) {
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/dns/1001/record" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"statusCode": 200,
			"id":         11,
			"domainId":   1001,
			"domainName": "example.com",
			"nodeName":   "alias",
			"hostname":   "alias.example.com",
			"recordType": "CNAME",
			"content":    "target.example.com",
			"host":       "target.example.com",
			"ttl":        90,
			"state":      true,
		})
	}))
	defer server.Close()

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(server.URL), dynuclient.WithHTTPClient(server.Client()))
	_, err := client.CreateDNSRecord(context.Background(), 1001, dynuclient.CreateDNSRecordRequest{
		NodeName:   "alias",
		RecordType: "CNAME",
		Content:    stringPointer("target.example.com"),
		TTL:        90,
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord() error = %v", err)
	}

	if captured["host"] != "target.example.com" {
		t.Fatalf("expected host in payload, got %#v", captured)
	}
	if _, ok := captured["content"]; ok {
		t.Fatalf("expected content to be omitted for CNAME record, got %#v", captured["content"])
	}
}

func TestClientCreateDNSRecordSendsContentForTXTRecord(t *testing.T) {
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/dns/1001/record" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"statusCode": 200,
			"id":         12,
			"domainId":   1001,
			"domainName": "example.com",
			"nodeName":   "txt",
			"hostname":   "txt.example.com",
			"recordType": "TXT",
			"content":    "hello",
			"ttl":        90,
			"state":      true,
		})
	}))
	defer server.Close()

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(server.URL), dynuclient.WithHTTPClient(server.Client()))
	_, err := client.CreateDNSRecord(context.Background(), 1001, dynuclient.CreateDNSRecordRequest{
		NodeName:   "txt",
		RecordType: "TXT",
		Content:    stringPointer("hello"),
		TTL:        90,
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord() error = %v", err)
	}

	if captured["content"] != "hello" {
		t.Fatalf("expected content in payload, got %#v", captured)
	}
	if _, ok := captured["ipv4Address"]; ok {
		t.Fatalf("expected ipv4Address to be omitted for TXT record, got %#v", captured["ipv4Address"])
	}
	if _, ok := captured["ipv6Address"]; ok {
		t.Fatalf("expected ipv6Address to be omitted for TXT record, got %#v", captured["ipv6Address"])
	}
}

func TestClientCreateDNSRecordNormalizesZoneStyleContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/dns/1001/record" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"statusCode": 200,
			"id":         10,
			"domainId":   1001,
			"domainName": "example.com",
			"nodeName":   "www",
			"hostname":   "www.example.com",
			"recordType": "A",
			"content":    "www.example.com. 300 IN A 167.179.167.166",
			"ttl":        300,
			"state":      true,
		})
	}))
	defer server.Close()

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(server.URL), dynuclient.WithHTTPClient(server.Client()))
	record, err := client.CreateDNSRecord(context.Background(), 1001, dynuclient.CreateDNSRecordRequest{
		NodeName:   "www",
		RecordType: "A",
		Content:    stringPointer("167.179.167.166"),
		TTL:        300,
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord() error = %v", err)
	}

	if record.Content != "167.179.167.166" {
		t.Fatalf("expected normalized content, got %q", record.Content)
	}
}

func TestClientDoRequestTopLevelAPIExceptionPayload(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()
	fake.SetRawResponse("/dns", 505, `{"statusCode":505,"type":"Validation Exception","message":"Invalid IP address."}`)

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	_, err := client.ListDomains(context.Background())
	if err == nil || !strings.Contains(err.Error(), "Invalid IP address.") {
		t.Fatalf("expected top-level API error, got %v", err)
	}
}

func TestClientCreateDNSRecordOmitsContentWhenUnset(t *testing.T) {
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/dns/1001/record" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"statusCode": 200,
			"id":         10,
			"domainId":   1001,
			"domainName": "example.com",
			"nodeName":   "www",
			"hostname":   "www.example.com",
			"recordType": "A",
			"ttl":        300,
			"state":      true,
		})
	}))
	defer server.Close()

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(server.URL), dynuclient.WithHTTPClient(server.Client()))
	_, err := client.CreateDNSRecord(context.Background(), 1001, dynuclient.CreateDNSRecordRequest{
		NodeName:   "www",
		RecordType: "A",
		TTL:        300,
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord() error = %v", err)
	}

	if _, ok := captured["content"]; ok {
		t.Fatalf("expected content to be omitted, got %#v", captured["content"])
	}
	if _, ok := captured["ipv4Address"]; ok {
		t.Fatalf("expected ipv4Address to be omitted, got %#v", captured["ipv4Address"])
	}
}

func stringPointer(value string) *string {
	return &value
}
