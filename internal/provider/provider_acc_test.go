package provider

import (
	"context"
	"os"
	"testing"

	"github.com/dynu/terraform-provider-dynu/internal/dynuclient"
)

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
