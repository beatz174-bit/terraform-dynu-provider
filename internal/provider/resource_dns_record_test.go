package provider

import "testing"

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

func TestHostnameMatchesDomain(t *testing.T) {
	testCases := []struct {
		name      string
		hostname  string
		domain    string
		wantMatch bool
	}{
		{name: "exact domain", hostname: "example.com", domain: "example.com", wantMatch: true},
		{name: "subdomain", hostname: "www.example.com", domain: "example.com", wantMatch: true},
		{name: "different root domain", hostname: "www.other.com", domain: "example.com", wantMatch: false},
		{name: "partial suffix does not match", hostname: "badexample.com", domain: "example.com", wantMatch: false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := hostnameMatchesDomain(tc.hostname, tc.domain)
			if got != tc.wantMatch {
				t.Fatalf("hostnameMatchesDomain(%q, %q) = %t, want %t", tc.hostname, tc.domain, got, tc.wantMatch)
			}
		})
	}
}
