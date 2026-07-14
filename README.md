# rw-swe-design-interview

A small Go service used for a **code-review + system-design exercise**. The
scenario is **cluster scaling** for a cloud service that runs a managed streaming
database (RisingWave) per customer: a customer resizes their cluster's compute
pool, and the control plane resizes the pods and rebalances the running streaming
jobs onto the new topology.

The implementation lives in `scaling/`. Your job is to review it and
discuss how you'd redesign it.

## Running the code

```bash
GOWORK=off go test ./...
```

There's a happy-path test that passes. See [`scaling/`](scaling/)
for the exercise itself.

---

*This repo holds only the exercise materials. The interviewer facilitation notes
and reference answers live in a separate private repo.*
