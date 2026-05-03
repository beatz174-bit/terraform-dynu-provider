package fakedynu

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"time"

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
	nextIDs    map[int64]int64
}

type rawResponse struct {
	httpStatus int
	body       string
}

type dnsRecordUpsertRequest struct {
	NodeName    string `json:"nodeName"`
	RecordType  string `json:"recordType"`
	Content     string `json:"content"`
	IPv4Address string `json:"ipv4Address"`
	IPv6Address string `json:"ipv6Address"`
	TTL         int64  `json:"ttl"`
	State       *bool  `json:"state"`
	Group       string `json:"group"`
	Host        string `json:"host"`
	Priority    int64  `json:"priority"`
	Weight      int64  `json:"weight"`
	Port        int64  `json:"port"`
	Flags       int64  `json:"flags"`
	Tag         string `json:"tag"`
	Value       string `json:"value"`
}
type domainUpsertRequest struct {
	Name        string `json:"name"`
	IPv4Address string `json:"ipv4Address"`
	IPv6Address string `json:"ipv6Address"`
	TTL         int64  `json:"ttl"`
	Group       string `json:"group"`
}

func NewServer() *Server {
	s := &Server{
		errors:     map[string]APIError{},
		rawPayload: map[string]rawResponse{},
		nextIDs:    map[int64]int64{},
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

	s.reseedNextIDs()
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
	s.reseedNextIDs()
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

func (s *Server) DeleteRecord(domainID int64, recordID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	records := s.fixture.RecordsByDomain[domainID]
	updated := make([]dynuclient.DNSRecord, 0, len(records))
	for _, record := range records {
		if record.ID != recordID {
			updated = append(updated, record)
		}
	}
	s.fixture.RecordsByDomain[domainID] = updated
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	case r.Method == http.MethodGet && r.URL.Path == "/dns":
		s.writeJSON(w, http.StatusOK, map[string]any{"statusCode": 200, "domains": s.fixture.Domains})
		return
	case r.Method == http.MethodPost && r.URL.Path == "/dns":
		s.handleCreateDomain(w, r)
		return
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/dns/getroot/"):
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
	case strings.HasPrefix(r.URL.Path, "/dns/"):
		s.serveDNSPath(w, r)
		return
	default:
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "endpoint not found"})
	}
}

func (s *Server) serveDNSPath(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/dns/")
	segments := strings.Split(trimmed, "/")
	if len(segments) == 1 {
		s.serveDomainByID(w, r, segments[0])
		return
	}

	if len(segments) >= 2 && segments[1] == "record" {
		s.serveRecordRoutes(w, r, segments)
		return
	}

	s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "endpoint not found"})
}

func (s *Server) serveDomainByID(w http.ResponseWriter, r *http.Request, rawDomainID string) {
	domainID, err := strconv.ParseInt(rawDomainID, 10, 64)
	if err != nil {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusBadRequest, StatusCode: 400, Type: "Validation Exception", Message: "invalid domain id: " + rawDomainID})
		return
	}
	for idx, domain := range s.fixture.Domains {
		if domain.ID == domainID {
			if r.Method == http.MethodDelete {
				s.fixture.Domains = append(s.fixture.Domains[:idx], s.fixture.Domains[idx+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if r.Method == http.MethodPost {
				s.handleUpdateDomain(w, r, idx)
				return
			}
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			payload := map[string]any{"statusCode": 200}
			b, _ := json.Marshal(domain)
			_ = json.Unmarshal(b, &payload)
			s.writeJSON(w, http.StatusOK, payload)
			return
		}
	}
	s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "domain not found"})
}

func (s *Server) handleCreateDomain(w http.ResponseWriter, r *http.Request) {
	var req domainUpsertRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusBadRequest, StatusCode: 400, Type: "Validation Exception", Message: "invalid request body"})
		return
	}
	id := int64(len(s.fixture.Domains) + 5000)
	domain := dynuclient.Domain{ID: id, Name: req.Name, UnicodeName: req.Name, IPv4Address: req.IPv4Address, IPv6Address: req.IPv6Address, TTL: req.TTL, Group: req.Group, State: "active", Token: fmt.Sprintf("tok-%d", id)}
	s.fixture.Domains = append(s.fixture.Domains, domain)
	payload := map[string]any{"statusCode": 200}
	b, _ := json.Marshal(domain)
	_ = json.Unmarshal(b, &payload)
	s.writeJSON(w, http.StatusOK, payload)
}
func (s *Server) handleUpdateDomain(w http.ResponseWriter, r *http.Request, idx int) {
	var req domainUpsertRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusBadRequest, StatusCode: 400, Type: "Validation Exception", Message: "invalid request body"})
		return
	}
	domain := s.fixture.Domains[idx]
	domain.Name = req.Name
	domain.UnicodeName = req.Name
	domain.IPv4Address = req.IPv4Address
	domain.IPv6Address = req.IPv6Address
	if req.TTL != 0 {
		domain.TTL = req.TTL
	}
	domain.Group = req.Group
	s.fixture.Domains[idx] = domain
	payload := map[string]any{"statusCode": 200}
	b, _ := json.Marshal(domain)
	_ = json.Unmarshal(b, &payload)
	s.writeJSON(w, http.StatusOK, payload)
}

func (s *Server) serveRecordRoutes(w http.ResponseWriter, r *http.Request, segments []string) {
	domainID, err := strconv.ParseInt(segments[0], 10, 64)
	if err != nil {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusBadRequest, StatusCode: 400, Type: "Validation Exception", Message: "invalid domain id: " + segments[0]})
		return
	}

	if len(segments) == 2 {
		switch r.Method {
		case http.MethodGet:
			records := s.fixture.RecordsByDomain[domainID]
			if records == nil {
				records = []dynuclient.DNSRecord{}
			}
			s.writeJSON(w, http.StatusOK, map[string]any{"statusCode": 200, "dnsRecords": records})
			return
		case http.MethodPost:
			s.handleCreateRecord(w, r, domainID)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

	if len(segments) == 3 {
		recordID, err := strconv.ParseInt(segments[2], 10, 64)
		if err != nil {
			s.writeAPIError(w, APIError{HTTPStatus: http.StatusBadRequest, StatusCode: 400, Type: "Validation Exception", Message: "invalid record id: " + segments[2]})
			return
		}
		switch r.Method {
		case http.MethodGet:
			s.handleGetRecord(w, domainID, recordID)
			return
		case http.MethodPut, http.MethodPost:
			s.handleUpdateRecord(w, r, domainID, recordID)
			return
		case http.MethodDelete:
			s.handleDeleteRecord(w, domainID, recordID)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

	s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "endpoint not found"})
}

func (s *Server) handleGetRecord(w http.ResponseWriter, domainID int64, recordID int64) {
	record, _, ok := s.findRecord(domainID, recordID)
	if !ok {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "record not found"})
		return
	}
	s.writeRecord(w, record)
}

func (s *Server) handleCreateRecord(w http.ResponseWriter, r *http.Request, domainID int64) {
	req, ok := s.decodeUpsertRequest(w, r)
	if !ok {
		return
	}
	if req.RecordType == "" {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusBadRequest, StatusCode: 400, Type: "Validation Exception", Message: "recordType is required"})
		return
	}
	content := contentFromUpsertRequest(req)
	if content == "" && !strings.EqualFold(req.RecordType, "A") && !strings.EqualFold(req.RecordType, "AAAA") {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusBadRequest, StatusCode: 400, Type: "Validation Exception", Message: "recordType and content are required"})
		return
	}

	domainName := s.domainName(domainID)
	if domainName == "" {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "domain not found"})
		return
	}
	state := true
	if req.State != nil {
		state = *req.State
	}
	ttl := req.TTL
	if ttl <= 0 {
		ttl = 90
	}
	record := dynuclient.DNSRecord{
		ID:         s.nextRecordID(domainID),
		DomainID:   domainID,
		DomainName: domainName,
		NodeName:   req.NodeName,
		Hostname:   buildHostname(req.NodeName, domainName),
		RecordType: req.RecordType,
		State:      state,
		TTL:        ttl,
		Content:    content,
		UpdatedOn:  time.Now().UTC().Format(time.RFC3339),
		Group:      req.Group,
		Host:       req.Host,
		Priority:   req.Priority,
		Weight:     req.Weight,
		Port:       req.Port,
		Flags:      req.Flags,
		Tag:        req.Tag,
		Value:      req.Value,
	}

	s.fixture.RecordsByDomain[domainID] = append(s.fixture.RecordsByDomain[domainID], record)
	s.writeRecord(w, record)
}

func (s *Server) handleUpdateRecord(w http.ResponseWriter, r *http.Request, domainID int64, recordID int64) {
	req, ok := s.decodeUpsertRequest(w, r)
	if !ok {
		return
	}
	record, idx, found := s.findRecord(domainID, recordID)
	if !found {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "record not found"})
		return
	}

	content := contentFromUpsertRequest(req)

	record.NodeName = req.NodeName
	record.RecordType = req.RecordType
	record.Content = content
	if req.TTL > 0 {
		record.TTL = req.TTL
	}
	if req.State != nil {
		record.State = *req.State
	}
	record.Group = req.Group
	record.Host = req.Host
	record.Priority = req.Priority
	record.Weight = req.Weight
	record.Port = req.Port
	record.Flags = req.Flags
	record.Tag = req.Tag
	record.Value = req.Value
	record.Hostname = buildHostname(req.NodeName, record.DomainName)
	record.UpdatedOn = time.Now().UTC().Format(time.RFC3339)

	s.fixture.RecordsByDomain[domainID][idx] = record
	s.writeRecord(w, record)
}

func contentFromUpsertRequest(req dnsRecordUpsertRequest) string {
	switch strings.ToUpper(strings.TrimSpace(req.RecordType)) {
	case "A":
		return strings.TrimSpace(req.IPv4Address)
	case "AAAA":
		return strings.TrimSpace(req.IPv6Address)
	case "CNAME":
		if strings.TrimSpace(req.Host) != "" {
			return strings.TrimSpace(req.Host)
		}
	case "MX", "SRV", "NS", "PTR":
		if strings.TrimSpace(req.Host) != "" {
			return strings.TrimSpace(req.Host)
		}
	case "CAA":
		if strings.TrimSpace(req.Value) != "" {
			return strings.TrimSpace(req.Value)
		}
	}
	return strings.TrimSpace(req.Content)
}

func (s *Server) handleDeleteRecord(w http.ResponseWriter, domainID int64, recordID int64) {
	_, idx, ok := s.findRecord(domainID, recordID)
	if !ok {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusNotFound, StatusCode: 404, Type: "Not Found", Message: "record not found"})
		return
	}

	records := s.fixture.RecordsByDomain[domainID]
	s.fixture.RecordsByDomain[domainID] = append(records[:idx], records[idx+1:]...)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) decodeUpsertRequest(w http.ResponseWriter, r *http.Request) (dnsRecordUpsertRequest, bool) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusBadRequest, StatusCode: 400, Type: "Validation Exception", Message: "unable to read request body"})
		return dnsRecordUpsertRequest{}, false
	}
	defer r.Body.Close()

	var req dnsRecordUpsertRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		s.writeAPIError(w, APIError{HTTPStatus: http.StatusBadRequest, StatusCode: 400, Type: "Validation Exception", Message: "invalid json payload"})
		return dnsRecordUpsertRequest{}, false
	}
	return req, true
}

func decodeJSONBody(body io.ReadCloser, target any) error {
	defer body.Close()
	payload, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, target)
}

func (s *Server) findRecord(domainID int64, recordID int64) (dynuclient.DNSRecord, int, bool) {
	records := s.fixture.RecordsByDomain[domainID]
	for idx, record := range records {
		if record.ID == recordID {
			return record, idx, true
		}
	}
	return dynuclient.DNSRecord{}, -1, false
}

func (s *Server) writeRecord(w http.ResponseWriter, record dynuclient.DNSRecord) {
	payload := map[string]any{"statusCode": 200}
	b, _ := json.Marshal(record)
	_ = json.Unmarshal(b, &payload)
	s.writeJSON(w, http.StatusOK, payload)
}

func (s *Server) domainName(domainID int64) string {
	for _, domain := range s.fixture.Domains {
		if domain.ID == domainID {
			return domain.Name
		}
	}
	return ""
}

func (s *Server) nextRecordID(domainID int64) int64 {
	nextID := s.nextIDs[domainID]
	if nextID == 0 {
		nextID = 1
	}
	s.nextIDs[domainID] = nextID + 1
	return nextID
}

func (s *Server) reseedNextIDs() {
	s.nextIDs = map[int64]int64{}
	for domainID, records := range s.fixture.RecordsByDomain {
		maxID := int64(0)
		for _, record := range records {
			if record.ID > maxID {
				maxID = record.ID
			}
		}
		s.nextIDs[domainID] = maxID + 1
	}
}

func buildHostname(nodeName string, domainName string) string {
	nodeName = strings.TrimSpace(nodeName)
	if nodeName == "" || nodeName == "@" {
		return domainName
	}
	return nodeName + "." + domainName
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
