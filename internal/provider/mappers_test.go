package provider

import (
	"testing"

	"github.com/dynu/terraform-provider-dynu/internal/dynuclient"
)

func TestSortDomains(t *testing.T) {
	domains := []dynuclient.Domain{
		{ID: 2, Name: "z.example.com"},
		{ID: 1, Name: "a.example.com"},
		{ID: 3, Name: "a.example.com"},
	}

	sortDomains(domains)

	if domains[0].ID != 1 || domains[1].ID != 3 || domains[2].ID != 2 {
		t.Fatalf("unexpected domain sort order: %#v", domains)
	}
}

func TestSortDNSRecords(t *testing.T) {
	records := []dynuclient.DNSRecord{
		{ID: 3, Hostname: "b.example.com", RecordType: "TXT"},
		{ID: 1, Hostname: "a.example.com", RecordType: "A"},
		{ID: 2, Hostname: "a.example.com", RecordType: "A"},
	}

	sortDNSRecords(records)

	if records[0].ID != 1 || records[1].ID != 2 || records[2].ID != 3 {
		t.Fatalf("unexpected record sort order: %#v", records)
	}
}

func TestMapString(t *testing.T) {
	if !mapString("").IsNull() {
		t.Fatal("expected empty string to map to null")
	}
	if mapString("value").ValueString() != "value" {
		t.Fatal("expected non-empty string to map to value")
	}
}
