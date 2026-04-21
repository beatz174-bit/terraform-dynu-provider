package dynuclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListDomains(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("API-Key"); got != "test-key" {
			t.Fatalf("unexpected api key header: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"statusCode":200,"domains":[{"id":1,"name":"example.com","state":"Complete"}]}`))
	}))
	defer ts.Close()

	client := New("test-key", WithBaseURL(ts.URL), WithHTTPClient(ts.Client()))
	domains, err := client.ListDomains(context.Background())
	if err != nil {
		t.Fatalf("ListDomains() error = %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(domains))
	}
	if domains[0].Name != "example.com" {
		t.Fatalf("unexpected domain name %q", domains[0].Name)
	}
}

func TestDoGetErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"statusCode":401,"exception":{"statusCode":401,"type":"Authentication Exception","message":"invalid"}}`))
	}))
	defer ts.Close()

	client := New("test-key", WithBaseURL(ts.URL), WithHTTPClient(ts.Client()))
	_, err := client.ListDomains(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
