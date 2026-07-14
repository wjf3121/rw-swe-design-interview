package scaling

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestServeHTTP is a happy-path smoke test: scaling an existing cluster up with
// fakes that always succeed returns 200 and leaves the cluster "running" at the
// new size.
//
// It intentionally does NOT exercise any of the failure modes the review is
// about (crashes, retries, timeouts, concurrent/conflicting scale requests,
// scale-down). Green here does not mean the design is sound.
func TestServeHTTP(t *testing.T) {
	store := NewInMemoryClusterStore(Cluster{ID: "acme", ComputeReplicas: 3, Status: "running"})
	s := &Scaler{
		DB:     store,
		K8s:    NewFakeK8s(),
		Stream: FakeStream{},
	}

	req := httptest.NewRequest(http.MethodPost, "/clusters/acme/scale",
		strings.NewReader(`{"cluster_id":"acme","compute_replicas":5}`))
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	got := store.Cluster("acme")
	if got.Status != "running" {
		t.Errorf("cluster status = %q, want %q", got.Status, "running")
	}
	if got.ComputeReplicas != 5 {
		t.Errorf("cluster replicas = %d, want 5", got.ComputeReplicas)
	}
}
