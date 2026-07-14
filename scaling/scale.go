// Package scaling implements cluster scaling for RisingWave Cloud.
//
// A customer changes the size of their cluster's compute pool; the control
// plane resizes the underlying pods and rebalances streaming jobs onto the new
// topology.
package scaling

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Cluster is the persisted record of a customer's cluster.
type Cluster struct {
	ID              string
	ComputeReplicas int
	Status          string
}

// ClusterStore persists cluster records.
type ClusterStore interface {
	Get(ctx context.Context, id string) (Cluster, error)
	Save(ctx context.Context, c Cluster) error
	SetStatus(ctx context.Context, id, status string) error
}

// K8sClient manages the compute pods backing a cluster.
type K8sClient interface {
	SetComputeReplicas(ctx context.Context, id string, replicas int) error
	PodsReady(ctx context.Context, id string) (bool, error)
}

// StreamManager reschedules streaming jobs across compute nodes.
type StreamManager interface {
	// Rebalance moves running stream actors so they spread across the current
	// set of compute nodes. It is expensive and disruptive: actors are paused
	// and their state is moved.
	Rebalance(ctx context.Context, id string) error
}

// Scaler wires together the dependencies needed to resize a cluster.
type Scaler struct {
	DB     ClusterStore
	K8s    K8sClient
	Stream StreamManager
}

type scaleRequest struct {
	ClusterID       string `json:"cluster_id"`
	ComputeReplicas int    `json:"compute_replicas"`
}

// ServeHTTP handles a scale request: it resizes a cluster's compute pool and
// rebalances its streaming jobs onto the new topology.
//
// Steps:
//  1. load the current cluster record
//  2. mark the cluster "scaling"
//  3. resize the compute pool
//  4. wait for the new pods to be ready
//  5. rebalance streaming jobs onto the new topology
//  6. persist the new size and mark the cluster "running"
func (s *Scaler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req scaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 1. Load the current state.
	cluster, err := s.DB.Get(ctx, req.ClusterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Mark the cluster as scaling.
	s.DB.SetStatus(ctx, req.ClusterID, "scaling")

	// 3. Resize the compute pool.
	if err := s.K8s.SetComputeReplicas(ctx, req.ClusterID, req.ComputeReplicas); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Wait for the new pods to come up.
	for {
		ready, err := s.K8s.PodsReady(ctx, req.ClusterID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if ready {
			break
		}
		time.Sleep(5 * time.Second)
	}

	// 5. Rebalance streaming jobs onto the new topology.
	// Rebalance is a heavy operation in RisingWave: it pauses running stream
	// actors and moves their state across nodes. It runs unconditionally on
	// every scale request, even when the cluster is already balanced for the
	// target size.
	if err := s.Stream.Rebalance(ctx, req.ClusterID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Persist the new size.
	cluster.ComputeReplicas = req.ComputeReplicas
	cluster.Status = "running"
	if err := s.DB.Save(ctx, cluster); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "cluster %s scaled to %d compute nodes\n", req.ClusterID, req.ComputeReplicas)
}
