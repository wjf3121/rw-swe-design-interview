# Code review + design exercise

This package implements **cluster scaling** for a cloud service that runs a
managed streaming database (RisingWave) for each customer. A customer can change
the size of their cluster's compute pool; the control plane resizes the pods and
rebalances the running streaming jobs onto the new set of nodes.

The whole thing lives in `scale.go`. `POST /clusters/{id}/scale` with a body
like:

```json
{ "cluster_id": "acme", "compute_replicas": 5 }
```

runs these steps:

1. load the current cluster record
2. mark the cluster `scaling`
3. resize the compute pool (add/remove pods)
4. wait for the new pods to be ready
5. rebalance streaming jobs onto the new topology
6. persist the new size and mark the cluster `running`

`infra.go` has in-memory fakes so it compiles and the test runs. There is a
happy-path test in `scale_test.go`:

```bash
GOWORK=off go test ./...
```

## Your task

Read `scale.go` and talk us through it as if it had landed in your review queue:
what you'd flag, and how you'd redesign it to be fault-tolerant. There's no need
to write code — just review and discuss.
