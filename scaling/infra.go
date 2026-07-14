package scaling

import (
	"context"
	"fmt"
	"sync"
)

// The types below are simple in-memory fakes so the service compiles and the
// happy-path test runs without a real database or cluster.

// InMemoryClusterStore is a trivial ClusterStore backed by a map.
type InMemoryClusterStore struct {
	mu       sync.Mutex
	clusters map[string]Cluster
}

func NewInMemoryClusterStore(seed ...Cluster) *InMemoryClusterStore {
	s := &InMemoryClusterStore{clusters: map[string]Cluster{}}
	for _, c := range seed {
		s.clusters[c.ID] = c
	}
	return s
}

func (s *InMemoryClusterStore) Get(_ context.Context, id string) (Cluster, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.clusters[id]
	if !ok {
		return Cluster{}, fmt.Errorf("cluster %s not found", id)
	}
	return c, nil
}

func (s *InMemoryClusterStore) Save(_ context.Context, c Cluster) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clusters[c.ID] = c
	return nil
}

func (s *InMemoryClusterStore) SetStatus(_ context.Context, id, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.clusters[id]
	c.Status = status
	s.clusters[id] = c
	return nil
}

func (s *InMemoryClusterStore) Cluster(id string) Cluster {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clusters[id]
}

// FakeK8s pretends to resize the compute pool and reports pods ready
// immediately.
type FakeK8s struct {
	mu       sync.Mutex
	replicas map[string]int
}

func NewFakeK8s() *FakeK8s {
	return &FakeK8s{replicas: map[string]int{}}
}

func (k *FakeK8s) SetComputeReplicas(_ context.Context, id string, replicas int) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.replicas[id] = replicas
	return nil
}

func (k *FakeK8s) PodsReady(_ context.Context, _ string) (bool, error) {
	return true, nil
}

// FakeStream pretends to rebalance streaming jobs.
type FakeStream struct{}

func (FakeStream) Rebalance(_ context.Context, _ string) error { return nil }
