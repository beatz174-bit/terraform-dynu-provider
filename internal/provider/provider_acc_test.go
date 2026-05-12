package provider

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/beatz174-bit/terraform-provider-dynu/internal/dynuclient"
)

func stringPtr(s string) *string {
	return &s
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}
	if os.Getenv("DYNU_API_KEY") == "" {
		t.Skip("set DYNU_API_KEY to run acceptance tests")
	}
}

func testAccDomainFromEnv(t *testing.T) string {
	t.Helper()
	domain := os.Getenv("DYNU_DOMAIN")
	if domain == "" {
		t.Skip("set DYNU_DOMAIN to run domain-specific acceptance tests")
	}
	return domain
}

func TestAccDataSourceDomains(t *testing.T) {
	testAccPreCheck(t)

	client := dynuclient.New(os.Getenv("DYNU_API_KEY"))
	domains, err := client.ListDomains(context.Background())
	if err != nil {
		t.Fatalf("ListDomains() failed: %v", err)
	}
	if len(domains) == 0 {
		t.Fatal("expected at least one domain")
	}
}

func TestAccDataSourceDomain(t *testing.T) {
	testAccPreCheck(t)
	hostname := testAccDomainFromEnv(t)

	client := dynuclient.New(os.Getenv("DYNU_API_KEY"))
	domainID, _, err := client.GetRootDomain(context.Background(), hostname)
	if err != nil {
		t.Fatalf("GetRootDomain() failed: %v", err)
	}
	domain, err := client.GetDomainByID(context.Background(), domainID)
	if err != nil {
		t.Fatalf("GetDomainByID() failed: %v", err)
	}
	if domain.ID == 0 || domain.Name == "" {
		t.Fatalf("unexpected domain payload: %#v", domain)
	}
}

func TestAccDataSourceDNSRecords(t *testing.T) {
	testAccPreCheck(t)
	hostname := testAccDomainFromEnv(t)

	client := dynuclient.New(os.Getenv("DYNU_API_KEY"))
	domainID, _, err := client.GetRootDomain(context.Background(), hostname)
	if err != nil {
		t.Fatalf("GetRootDomain() failed: %v", err)
	}
	records, err := client.ListDNSRecords(context.Background(), domainID)
	if err != nil {
		t.Fatalf("ListDNSRecords() failed: %v", err)
	}
	if records == nil {
		t.Fatal("expected records slice, got nil")
	}
}

func TestAccDNSRecordAWithEmptyContent(t *testing.T) {
	testAccPreCheck(t)
	hostname := testAccDomainFromEnv(t)
	domainID, _, client := testAccDomainClient(t, hostname)

	nodeName := testAccDisposableNodeName("acc-empty-a")
	record := testAccCreateRecordMaybeSkipUnsupported(
		t,
		client,
		domainID,
		dynuclient.CreateDNSRecordRequest{
			NodeName:   nodeName,
			RecordType: "A",
		},
		"A record with empty content",
	)
	defer testAccDeleteRecord(t, client, domainID, record.ID)

	got := testAccFetchRecordFromList(t, client, domainID, record.ID)
	if strings.TrimSpace(got.Content) != "" {
		t.Fatalf("expected Dynu to return empty content for A record, got %q", got.Content)
	}
}

func TestAccDNSRecordAAAAWithEmptyContent(t *testing.T) {
	testAccPreCheck(t)
	hostname := testAccDomainFromEnv(t)
	domainID, _, client := testAccDomainClient(t, hostname)

	nodeName := testAccDisposableNodeName("acc-empty-aaaa")
	record := testAccCreateRecordMaybeSkipUnsupported(
		t,
		client,
		domainID,
		dynuclient.CreateDNSRecordRequest{
			NodeName:   nodeName,
			RecordType: "AAAA",
		},
		"AAAA record with empty content",
	)
	defer testAccDeleteRecord(t, client, domainID, record.ID)

	got := testAccFetchRecordFromList(t, client, domainID, record.ID)
	if strings.TrimSpace(got.Content) != "" {
		t.Fatalf("expected Dynu to return empty content for AAAA record, got %q", got.Content)
	}
}

func TestAccDNSRecordCNAMELifecycle(t *testing.T) {
	testAccPreCheck(t)
	hostname := testAccDomainFromEnv(t)
	domainID, _, client := testAccDomainClient(t, hostname)

	nodeName := testAccDisposableNodeName("acc-cname")

	created, err := client.CreateDNSRecord(context.Background(), domainID, dynuclient.CreateDNSRecordRequest{
		NodeName:   nodeName,
		RecordType: "CNAME",
		Content:    stringPtr("target1.example.com"),
		TTL:        120,
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord() failed for CNAME create: %v", err)
	}
	defer testAccDeleteRecord(t, client, domainID, created.ID)

	updated, err := client.UpdateDNSRecord(context.Background(), domainID, created.ID, dynuclient.UpdateDNSRecordRequest{
		NodeName:   nodeName,
		RecordType: "CNAME",
		Content:    stringPtr("target2.example.com"),
		TTL:        300,
	})
	if err != nil {
		t.Fatalf("UpdateDNSRecord() failed for CNAME update: %v", err)
	}
	if got := strings.TrimSuffix(strings.ToLower(updated.Content), "."); got != "target2.example.com" {
		t.Fatalf("unexpected updated content from create/update response: %q", updated.Content)
	}

	read := testAccFetchRecordFromList(t, client, domainID, created.ID)
	if got := strings.TrimSuffix(strings.ToLower(read.Content), "."); got != "target2.example.com" {
		t.Fatalf("expected ListDNSRecords() read-back to reflect updated CNAME target, got %q", read.Content)
	}
}

func testAccDomainClient(t *testing.T, hostname string) (int64, string, *dynuclient.Client) {
	t.Helper()
	client := dynuclient.New(os.Getenv("DYNU_API_KEY"))
	domainID, domainName, err := client.GetRootDomain(context.Background(), hostname)
	if err != nil {
		t.Fatalf("GetRootDomain() failed: %v", err)
	}
	return domainID, domainName, client
}

func testAccDisposableNodeName(prefix string) string {
	return fmt.Sprintf("%s-%d-%04d", prefix, time.Now().UnixNano(), rand.Intn(10000))
}

func testAccCreateRecordMaybeSkipUnsupported(t *testing.T, client *dynuclient.Client, domainID int64, req dynuclient.CreateDNSRecordRequest, scenario string) *dynuclient.DNSRecord {
	t.Helper()
	record, err := client.CreateDNSRecord(context.Background(), domainID, req)
	if err == nil {
		return record
	}

	var apiErr *dynuclient.APIError
	if errors.As(err, &apiErr) && isUnsupportedEmptyContentError(apiErr) {
		t.Skipf("Dynu account/API does not support %s in this environment (%v)", scenario, err)
	}

	t.Fatalf("CreateDNSRecord() failed for %s: %v", scenario, err)
	return nil
}

func testAccFetchRecordFromList(t *testing.T, client *dynuclient.Client, domainID int64, recordID int64) dynuclient.DNSRecord {
	t.Helper()
	records, err := client.ListDNSRecords(context.Background(), domainID)
	if err != nil {
		t.Fatalf("ListDNSRecords() failed: %v", err)
	}
	for _, record := range records {
		if record.ID == recordID {
			return record
		}
	}
	t.Fatalf("record id %d not found in ListDNSRecords() response", recordID)
	return dynuclient.DNSRecord{}
}

func testAccDeleteRecord(t *testing.T, client *dynuclient.Client, domainID int64, recordID int64) {
	t.Helper()
	if err := client.DeleteDNSRecord(context.Background(), domainID, recordID); err != nil {
		t.Fatalf("DeleteDNSRecord() cleanup failed: %v", err)
	}
}

func TestIsUnsupportedEmptyContentAPIError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		err    *dynuclient.APIError
		expect bool
	}{
		{
			name: "validation message with required content",
			err: &dynuclient.APIError{
				StatusCode: 400,
				Type:       "Validation Exception",
				Message:    "Content is required.",
			},
			expect: true,
		},
		{
			name: "validation message with ipv4 required",
			err: &dynuclient.APIError{
				StatusCode: 400,
				Type:       "Validation Exception",
				Message:    "IPv4Address is required for A records.",
			},
			expect: true,
		},
		{
			name: "different validation error should fail",
			err: &dynuclient.APIError{
				StatusCode: 400,
				Type:       "Validation Exception",
				Message:    "recordType is invalid",
			},
			expect: false,
		},
		{
			name: "transient throttling should fail",
			err: &dynuclient.APIError{
				StatusCode: 429,
				Type:       "Too Many Requests",
				Message:    "rate limit exceeded",
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isUnsupportedEmptyContentError(tc.err); got != tc.expect {
				t.Fatalf("unexpected result for %q: got %v, want %v", tc.name, got, tc.expect)
			}
		})
	}
}
