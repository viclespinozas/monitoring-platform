package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// memRepo is an in-memory Repository for exercising the HTTP and service
// layers without a database. The SQL itself is covered by integration tests.
type memRepo struct {
	mu    sync.Mutex
	next  int
	items []Service
}

func (m *memRepo) find(id string) int {
	for i, s := range m.items {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func (m *memRepo) nameTaken(name, exceptID string) bool {
	for _, s := range m.items {
		if s.Name == name && s.ID != exceptID {
			return true
		}
	}
	return false
}

func (m *memRepo) Create(_ context.Context, in Input) (Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nameTaken(in.Name, "") {
		return Service{}, ErrNameTaken
	}
	m.next++
	now := time.Now().UTC()
	s := fromInput(fmt.Sprintf("id-%d", m.next), in)
	s.CreatedAt, s.UpdatedAt = now, now
	m.items = append(m.items, s)
	return s, nil
}

func (m *memRepo) Get(_ context.Context, id string) (Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if i := m.find(id); i >= 0 {
		return m.items[i], nil
	}
	return Service{}, ErrNotFound
}

func (m *memRepo) List(_ context.Context, f ListFilter) ([]Service, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var matched []Service
	for _, s := range m.items {
		if f.Team != "" && s.Team != f.Team {
			continue
		}
		if f.Tag != "" && !slices.Contains(s.Tags, f.Tag) {
			continue
		}
		matched = append(matched, s)
	}
	total := len(matched)
	if f.Offset >= total {
		return []Service{}, total, nil
	}
	end := min(f.Offset+f.Limit, total)
	return matched[f.Offset:end], total, nil
}

func (m *memRepo) Update(_ context.Context, id string, in Input) (Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i := m.find(id)
	if i < 0 {
		return Service{}, ErrNotFound
	}
	if m.nameTaken(in.Name, id) {
		return Service{}, ErrNameTaken
	}
	s := fromInput(id, in)
	s.CreatedAt, s.UpdatedAt = m.items[i].CreatedAt, time.Now().UTC()
	m.items[i] = s
	return s, nil
}

func (m *memRepo) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	i := m.find(id)
	if i < 0 {
		return ErrNotFound
	}
	m.items = slices.Delete(m.items, i, i+1)
	return nil
}

func fromInput(id string, in Input) Service {
	return Service{
		ID: id, Name: in.Name, Team: in.Team, Description: in.Description, Tags: in.Tags,
		HealthURL: in.HealthURL, CheckIntervalSeconds: in.CheckIntervalSeconds,
		TimeoutSeconds: in.TimeoutSeconds, ExpectedStatus: in.ExpectedStatus,
	}
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	NewHandler(NewManager(&memRepo{}), logger).Routes(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, method, url, body string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, b
}

const orderBody = `{"name":"orders","team":"core","tags":["prod"],"health_url":"http://orders/healthz"}`

func TestCreateGetUpdateDeleteFlow(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/services"

	resp, body := do(t, "POST", base, orderBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, body %s", resp.StatusCode, body)
	}
	var created Service
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.CheckIntervalSeconds != 30 {
		t.Fatalf("unexpected created service: %+v", created)
	}
	if loc := resp.Header.Get("Location"); loc != "/api/v1/services/"+created.ID {
		t.Errorf("Location = %q", loc)
	}

	resp, _ = do(t, "GET", base+"/"+created.ID, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d", resp.StatusCode)
	}

	updated := `{"name":"orders","team":"payments","health_url":"http://orders/healthz","check_interval_seconds":60}`
	resp, body = do(t, "PUT", base+"/"+created.ID, updated)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d, body %s", resp.StatusCode, body)
	}
	var got Service
	_ = json.Unmarshal(body, &got)
	if got.Team != "payments" || got.CheckIntervalSeconds != 60 {
		t.Fatalf("update not applied: %+v", got)
	}

	resp, _ = do(t, "DELETE", base+"/"+created.ID, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d", resp.StatusCode)
	}
	resp, _ = do(t, "GET", base+"/"+created.ID, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete status = %d", resp.StatusCode)
	}
	resp, _ = do(t, "DELETE", base+"/"+created.ID, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("second delete status = %d", resp.StatusCode)
	}
}

func TestErrorsUseProblemJSON(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/services"
	do(t, "POST", base, orderBody)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantFields bool
	}{
		{"validation", "POST", base, `{"name":"BAD","team":"","health_url":"x"}`, 422, true},
		{"duplicate name", "POST", base, orderBody, 409, false},
		{"not found", "GET", base + "/nope", "", 404, false},
		{"malformed json", "POST", base, `{"name":`, 400, false},
		{"unknown field", "POST", base, `{"name":"a","team":"t","health_url":"http://a","extra":1}`, 400, false},
		{"trailing data", "POST", base, orderBody + `{}`, 400, false},
		{"bad limit", "GET", base + "?limit=abc", "", 422, true},
		{"negative offset", "GET", base + "?offset=-1", "", 422, true},
		{"update missing", "PUT", base + "/nope", orderBody, 404, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := do(t, tt.method, tt.path, tt.body)
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body %s", resp.StatusCode, tt.wantStatus, body)
			}
			if ct := resp.Header.Get("Content-Type"); ct != "application/problem+json" {
				t.Fatalf("Content-Type = %q", ct)
			}
			var p Problem
			if err := json.Unmarshal(body, &p); err != nil {
				t.Fatal(err)
			}
			if p.Status != tt.wantStatus || p.Title == "" || p.Type == "" {
				t.Fatalf("bad problem: %+v", p)
			}
			if tt.wantFields && len(p.Errors) == 0 {
				t.Fatalf("expected field errors: %+v", p)
			}
		})
	}
}

func TestListFilteringAndPaging(t *testing.T) {
	srv := newTestServer(t)
	base := srv.URL + "/api/v1/services"
	for i, spec := range []struct{ team, tag string }{
		{"core", "prod"}, {"core", "dev"}, {"payments", "prod"}, {"core", "prod"},
	} {
		body := fmt.Sprintf(`{"name":"svc-%d","team":%q,"tags":[%q],"health_url":"http://x/h"}`, i, spec.team, spec.tag)
		if resp, b := do(t, "POST", base, body); resp.StatusCode != http.StatusCreated {
			t.Fatalf("seed failed: %s", b)
		}
	}

	list := func(query string) listResponse {
		t.Helper()
		resp, body := do(t, "GET", base+query, "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("list %q status = %d, body %s", query, resp.StatusCode, body)
		}
		var lr listResponse
		if err := json.Unmarshal(body, &lr); err != nil {
			t.Fatal(err)
		}
		return lr
	}

	if lr := list(""); lr.Total != 4 || len(lr.Items) != 4 || lr.Limit != DefaultPageSize {
		t.Errorf("unfiltered: %+v", lr)
	}
	if lr := list("?team=core"); lr.Total != 3 {
		t.Errorf("team filter total = %d", lr.Total)
	}
	if lr := list("?tag=prod"); lr.Total != 3 {
		t.Errorf("tag filter total = %d", lr.Total)
	}
	if lr := list("?team=core&tag=prod"); lr.Total != 2 {
		t.Errorf("combined filter total = %d", lr.Total)
	}
	if lr := list("?limit=2&offset=1"); lr.Total != 4 || len(lr.Items) != 2 || lr.Offset != 1 {
		t.Errorf("paging: %+v", lr)
	}
	if lr := list("?limit=1000"); lr.Limit != MaxPageSize {
		t.Errorf("limit not clamped: %d", lr.Limit)
	}
	if lr := list("?offset=99"); lr.Total != 4 || len(lr.Items) != 0 {
		t.Errorf("offset past end: %+v", lr)
	}
}
