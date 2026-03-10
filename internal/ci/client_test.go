package ci

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(server *httptest.Server) *Client {
	c := NewClient("test-token")
	c.baseURL = server.URL
	return c
}

func TestLatestPipeline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Circle-Token") != "test-token" {
			t.Error("missing auth token")
		}
		resp := PipelineResponse{
			Items: []Pipeline{
				{ID: "pipe-1", Number: 42, State: "created"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server)
	p, err := c.LatestPipeline("gh/org/repo", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != "pipe-1" {
		t.Errorf("ID = %q, want %q", p.ID, "pipe-1")
	}
	if p.Number != 42 {
		t.Errorf("Number = %d, want %d", p.Number, 42)
	}
}

func TestLatestPipelineNoPipelines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PipelineResponse{})
	}))
	defer server.Close()

	c := newTestClient(server)
	_, err := c.LatestPipeline("gh/org/repo", "main")
	if err == nil {
		t.Fatal("expected error for empty pipelines")
	}
}

func TestWorkflows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := WorkflowResponse{
			Items: []Workflow{
				{ID: "wf-1", Name: "build-and-release", Status: "success"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server)
	wfs, err := c.Workflows("pipe-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wfs) != 1 || wfs[0].Name != "build-and-release" {
		t.Errorf("unexpected workflows: %+v", wfs)
	}
}

func TestJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := JobResponse{
			Items: []Job{
				{ID: "job-1", Name: "lint-and-test", Status: "success", JobNumber: 100},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server)
	jobs, err := c.Jobs("wf-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 1 || jobs[0].Name != "lint-and-test" {
		t.Errorf("unexpected jobs: %+v", jobs)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"unauthorized"}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	_, err := c.LatestPipeline("gh/org/repo", "main")
	if err == nil {
		t.Fatal("expected error for 401")
	}
}
