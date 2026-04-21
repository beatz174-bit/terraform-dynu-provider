package dynuclient_test

import (
	"context"
	"github.com/dynu/terraform-provider-dynu/internal/dynuclient"
	"strings"
	"testing"

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

func TestClientDoGetNon2xxStatus(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()
	fake.SetRawResponse("/dns", 401, `{"message":"nope"}`)

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	_, err := client.ListDomains(context.Background())
	if err == nil || !strings.Contains(err.Error(), "status 401") {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestClientDoGetAPIExceptionPayload(t *testing.T) {
	fake := fakedynu.NewServer()
	defer fake.Close()
	fake.SetAPIError("/dns", fakedynu.APIError{HTTPStatus: 400, StatusCode: 400, Type: "Validation Exception", Message: "bad hostname"})

	client := dynuclient.New("test-key", dynuclient.WithBaseURL(fake.BaseURL()), dynuclient.WithHTTPClient(fake.Client()))
	_, err := client.ListDomains(context.Background())
	if err == nil || !strings.Contains(err.Error(), "Validation Exception") {
		t.Fatalf("expected API exception error, got %v", err)
	}
}

func TestClientDoGetMalformedJSON(t *testing.T) {
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
