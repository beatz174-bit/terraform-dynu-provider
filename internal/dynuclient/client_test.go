package dynuclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientListDomainsSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("API-Key"); got != "test-key" {
			t.Fatalf("unexpected api key header: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"statusCode":200,"domains":[{"id":2,"name":"z.example.com"},{"id":1,"name":"a.example.com"}]}`))
	}))
	defer ts.Close()

	client := New("test-key", WithBaseURL(ts.URL), WithHTTPClient(ts.Client()))
	domains, err := client.ListDomains(context.Background())
	if err != nil {
		t.Fatalf("ListDomains() error = %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(domains))
	}
}

func TestClientDoGetNon2xxStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"nope"}`))
	}))
	defer ts.Close()

	client := New("test-key", WithBaseURL(ts.URL), WithHTTPClient(ts.Client()))
	_, err := client.ListDomains(context.Background())
	if err == nil || !strings.Contains(err.Error(), "status 401") {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestClientDoGetAPIExceptionPayload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"statusCode":200,"exception":{"statusCode":400,"type":"Validation Exception","message":"bad hostname"}}`))
	}))
	defer ts.Close()

	client := New("test-key", WithBaseURL(ts.URL), WithHTTPClient(ts.Client()))
	_, err := client.ListDomains(context.Background())
	if err == nil || !strings.Contains(err.Error(), "Validation Exception") {
		t.Fatalf("expected API exception error, got %v", err)
	}
}

func TestClientDoGetMalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"statusCode":200,"domains":[`))
	}))
	defer ts.Close()

	client := New("test-key", WithBaseURL(ts.URL), WithHTTPClient(ts.Client()))
	_, err := client.ListDomains(context.Background())
	if err == nil || !strings.Contains(err.Error(), "failed to decode dynu API response") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestClientGetRootDomainIncompleteResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"statusCode":200,"id":0,"domainName":""}`))
	}))
	defer ts.Close()

	client := New("test-key", WithBaseURL(ts.URL), WithHTTPClient(ts.Client()))
	_, _, err := client.GetRootDomain(context.Background(), "www.example.com")
	if err == nil || !strings.Contains(err.Error(), "incomplete root domain response") {
		t.Fatalf("expected incomplete response error, got %v", err)
	}
}

func TestClientGetRootDomainEscapesHostname(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.RequestURI, "spaces%20in%20name.example.com") {
			t.Fatalf("expected escaped hostname path, got %s", r.RequestURI)
		}
		_, _ = w.Write([]byte(`{"statusCode":200,"id":123,"domainName":"example.com"}`))
	}))
	defer ts.Close()

	client := New("test-key", WithBaseURL(ts.URL), WithHTTPClient(ts.Client()))
	_, _, err := client.GetRootDomain(context.Background(), "spaces in name.example.com")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
