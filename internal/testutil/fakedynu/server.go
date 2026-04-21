package fakedynu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"

	"github.com/dynu/terraform-provider-dynu/internal/dynuclient"
)

type RootDomain struct {
	ID         int64
	Hostname   string
	DomainName string
	Node       string
}

type APIError struct {
	HTTPStatus int
	StatusCode int
	Type       string
	Message    string
}

type Fixture struct {
	Domains         []dynuclient.Domain
	RootsByHostname map[string]RootDomain
	RecordsByDomain map[int64][]dynuclient.DNSRecord
}

type Server struct {
	*httptest.Server

	mu         sync.RWMutex
	fixture    Fixture
	errors     map[string]APIError
	rawPayload map[string]rawResponse
}

type rawResponse struct {
	httpStatus int
	body       string
}

func NewServer() *Server {
	s := &Server{
		errors:     map[string]APIError{},
		rawPayload: map[string]rawResponse{},
		fixture: Fixture{
			Domains: []dynuclient.Domain{
				{ID: 2002, Name: "z.example.com", UnicodeName: "z.example.com", TTL: 60, CreatedOn: "2024-01-03T00:00:00", UpdatedOn: "2024-01-04T00:00:00"},
				{ID: 1001, Name: "a.example.com", UnicodeName: "a.example.com", TTL: 120, CreatedOn: "2024-01-01T00:00:00", UpdatedOn: "2024-01-02T00:00:00"},
			},
			RootsByHostname: map[string]RootDomain{
				"www.a.example.com": {ID: 1001, Hostname: "www.a.example.com", DomainName: "a.example.com", Node: "www"},
				"api.a.example.com": {ID: 1001, Hostname: "api.a.example.com", DomainName: "a.example.com", Node: "api"},
			},
			RecordsByDomain: map[int64][]dynuclient.DNSRecord{
				1001: {
					{ID: 20, DomainID: 1001, DomainName: "a.example.com", NodeName: "www", Hostname: "www.a.example.com", RecordType: "TXT", TTL: 90, State: true, Content: "v=spf1", UpdatedOn: "2024-01-03T11:00:00"},
					{ID: 10, DomainID: 1001, DomainName: "a.example.com", NodeName: "www", Hostname: "www.a.example.com", RecordType: "A", TTL: 30, State: true, Content: "203.0.113.5", UpdatedOn: "2024-01-03T10:00:00"},
				},
			},
		},
	}

	s.Server = httptest.NewServer(http.HandlerFunc(s.serveHTTP))
	return s
}

func (s *Server) BaseURL() string {
	return s.URL
}

func (s *Server) SetFixture(fixture Fixture) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fixture = fixture
}

func (s *Server) SetAPIError(path string, apiErr APIError) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors[path] = apiErr
}

func (s *Server) SetRawResponse(path string, httpStatus int, body string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rawPayload[path] = rawResponse{httpStatus: httpStatus, body: body}
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if raw, ok := s.rawPayload[r.URL.Path]; ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(raw.httpStatus)
		_, _ = w.Write([]byte(raw.body))
		return
	}

	if apiErr, ok := s.errors[r.URL.Path]; ok {
		s.writeAPIError(w, apiErr)
		return
	}

	switch {
	case r.URL.Path == "/dns":
		s.writeJSON(w, http.StatusOK, map[string]any{"statusCode": 200, "domains": s.fixture.Domains})
		return
	case strings.HasPrefix(r.URL.Path, "/dns/getroot/"):
		hostname := strings.TrimPrefix(r.URL.Path, "/dns/getroot/")
		root, ok := s.fixture.RootsByHostname[hostname]
		if !ok {
			s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "hostname not found"})
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]any{
			"statusCode": 200,
			"id":         root.ID,
			"hostname":   root.Hostname,
			"domainName": root.DomainName,
			"node":       root.Node,
		})
		return
	case strings.HasSuffix(r.URL.Path, "/record"):
		domainID, err := domainIDFromPath(strings.TrimSuffix(r.URL.Path, "/record"))
		if err != nil {
			s.writeAPIError(w, APIError{HTTPStatus: http.StatusBadRequest, StatusCode: 400, Type: "Validation Exception", Message: err.Error()})
			return
		}
		records, ok := s.fixture.RecordsByDomain[domainID]
		if !ok {
			records = []dynuclient.DNSRecord{}
		}
		s.writeJSON(w, http.StatusOK, map[string]any{"statusCode": 200, "dnsRecords": records})
		return
	default:
		domainID, err := domainIDFromPath(r.URL.Path)
		if err != nil {
			s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "endpoint not found"})
			return
		}
		for _, domain := range s.fixture.Domains {
			if domain.ID == domainID {
				payload := map[string]any{"statusCode": 200}
				b, _ := json.Marshal(domain)
				_ = json.Unmarshal(b, &payload)
				s.writeJSON(w, http.StatusOK, payload)
				return
			}
		}
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "domain not found"})
	}
}

func domainIDFromPath(path string) (int64, error) {
	trimmed := strings.TrimPrefix(path, "/dns/")
	if trimmed == path {
		return 0, fmt.Errorf("invalid domain path: %s", path)
	}

	id, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid domain id: %s", trimmed)
	}
	return id, nil
}

func (s *Server) writeAPIError(w http.ResponseWriter, apiErr APIError) {
	if apiErr.HTTPStatus == 0 {
		apiErr.HTTPStatus = http.StatusBadRequest
	}
	if apiErr.StatusCode == 0 {
		apiErr.StatusCode = apiErr.HTTPStatus
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"statusCode": apiErr.StatusCode,
		"exception": map[string]any{
			"statusCode": apiErr.StatusCode,
			"type":       apiErr.Type,
			"message":    apiErr.Message,
		},
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
