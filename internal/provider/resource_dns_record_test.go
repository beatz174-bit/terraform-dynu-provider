package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	ipv4 := "192.0.2.123"
	ipv6 := "2001:db8::123"
	nonEmpty := "hello"
	blank := ""

	tests := []struct {
		name       string
		recordType string
		content    *string
		dynamic    bool
		wantValid  bool
	}{
		{name: "A accepts static ipv4", recordType: "A", content: &ipv4, wantValid: true},
		{name: "A rejects ipv6", recordType: "A", content: &ipv6, wantValid: false},
		{name: "A accepts dynamic nil", recordType: "A", content: nil, dynamic: true, wantValid: true},
		{name: "A accepts dynamic blank", recordType: "A", content: &blank, dynamic: true, wantValid: true},
		{name: "AAAA accepts static ipv6", recordType: "AAAA", content: &ipv6, wantValid: true},
		{name: "AAAA rejects ipv4", recordType: "AAAA", content: &ipv4, wantValid: false},
		{name: "AAAA accepts dynamic nil", recordType: "AAAA", content: nil, dynamic: true, wantValid: true},
		{name: "TXT requires content", recordType: "TXT", content: nil, wantValid: false},
		{name: "TXT rejects blank content", recordType: "TXT", content: &blank, wantValid: false},
		{name: "TXT accepts non-empty content", recordType: "TXT", content: &nonEmpty, wantValid: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diags := diag.Diagnostics{}
			got := validateDNSRecordContentForType(tc.recordType, tc.content, tc.dynamic, &diags)
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

func TestResolveDynamicIntent(t *testing.T) {
	diags := diag.Diagnostics{}
	dynamic, ok := resolveDynamicIntent("A", types.StringNull(), types.BoolNull(), &diags)
	if !ok || !dynamic || diags.HasError() {
		t.Fatalf("expected omitted A content to resolve to dynamic=true, got dynamic=%v ok=%v diags=%v", dynamic, ok, diags)
	}

	diags = diag.Diagnostics{}
	dynamic, ok = resolveDynamicIntent("A", types.StringValue("192.0.2.10"), types.BoolNull(), &diags)
	if !ok || dynamic || diags.HasError() {
		t.Fatalf("expected static A content to resolve to dynamic=false, got dynamic=%v ok=%v diags=%v", dynamic, ok, diags)
	}
}

func TestNormalizeRecordContentForState(t *testing.T) {
	if got := normalizeRecordContentForState("AAAA", "2001:0db8:0000:0000:0000:0000:0000:0123", false); got.ValueString() != "2001:db8::123" {
		t.Fatalf("expected canonical IPv6, got %q", got.ValueString())
	}
	if got := normalizeRecordContentForState("CNAME", "Example.COM.", false); got.ValueString() != "Example.COM" {
		t.Fatalf("expected trailing dot removed, got %q", got.ValueString())
	}
	if got := normalizeRecordContentForState("A", "(167.179.167.166)", true); !got.IsNull() {
		t.Fatalf("expected dynamic content to remain null, got %q", got.ValueString())
	}
}

func TestInferDynamicIntentFromState(t *testing.T) {
	tests := []struct {
		name       string
		recordType types.String
		content    types.String
		dynamic    types.Bool
		want       bool
	}{
		{
			name:       "explicit dynamic false wins",
			recordType: types.StringValue("A"),
			content:    types.StringNull(),
			dynamic:    types.BoolValue(false),
			want:       false,
		},
		{
			name:       "legacy A null content treated dynamic",
			recordType: types.StringValue("A"),
			content:    types.StringNull(),
			dynamic:    types.BoolNull(),
			want:       true,
		},
		{
			name:       "legacy AAAA unknown content treated dynamic",
			recordType: types.StringValue("AAAA"),
			content:    types.StringUnknown(),
			dynamic:    types.BoolNull(),
			want:       true,
		},
		{
			name:       "legacy non-A not dynamic",
			recordType: types.StringValue("TXT"),
			content:    types.StringNull(),
			dynamic:    types.BoolNull(),
			want:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := inferDynamicIntentFromState(tc.recordType, tc.content, tc.dynamic); got != tc.want {
				t.Fatalf("inferDynamicIntentFromState()=%v, want %v", got, tc.want)
			}
		})
	}
}
