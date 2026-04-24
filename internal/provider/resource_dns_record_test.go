package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestParseDNSRecordID(t *testing.T) {
	domainID, recordID, err := parseDNSRecordID("1001/55")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if domainID != 1001 || recordID != 55 {
		t.Fatalf("unexpected ids: domain=%d record=%d", domainID, recordID)
	}
}

func TestParseDNSRecordIDInvalid(t *testing.T) {
	if _, _, err := parseDNSRecordID("1001"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestValidateDNSRecordContentForType(t *testing.T) {
	nonEmpty := "hello"
	blank := ""

	tests := []struct {
		name       string
		recordType string
		content    *string
		wantValid  bool
	}{
		{name: "A allows nil content", recordType: "A", content: nil, wantValid: true},
		{name: "AAAA allows nil content", recordType: "AAAA", content: nil, wantValid: true},
		{name: "TXT requires content", recordType: "TXT", content: nil, wantValid: false},
		{name: "TXT rejects blank content", recordType: "TXT", content: &blank, wantValid: false},
		{name: "TXT accepts non-empty content", recordType: "TXT", content: &nonEmpty, wantValid: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diags := diag.Diagnostics{}
			got := validateDNSRecordContentForType(tc.recordType, tc.content, &diags)
			if got != tc.wantValid {
				t.Fatalf("validateDNSRecordContentForType()=%v, want %v", got, tc.wantValid)
			}
			if tc.wantValid && diags.HasError() {
				t.Fatalf("expected no error diagnostics, got %#v", diags)
			}
			if !tc.wantValid && !diags.HasError() {
				t.Fatal("expected error diagnostics")
			}
		})
	}
}
