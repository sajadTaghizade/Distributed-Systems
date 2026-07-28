# Replicated Key-Value Store (Distributed Systems — HW3)

A simple **replicated key-value store** that demonstrates `replication` and the
difference between **eventual** and **strong (majority-quorum)** consistency. The
cluster is three independent replica processes that communicate over HTTP, plus a
CLI client.

## Layout

```
.
├── replica/      # one cluster node (HTTP service) + README
├── client/       # CLI client + README
├── configs/      # replica1/2/3.json topology files
├── results/      # captured output of the four experiment scenarios
├── report.pdf    # full report (Persian)
└── go.mod
```

## Prerequisites

Go 1.21+ (`go version`).

## Build

```sh
go build -o bin/replica.exe ./replica
go build -o bin/client.exe  ./client
```

## Run the cluster

Start each replica in its own terminal (using the provided config files):

```sh
go run ./replica -config configs/replica1.json    # replica1 on :8081
go run ./replica -config configs/replica2.json    # replica2 on :8082
go run ./replica -config configs/replica3.json    # replica3 on :8083
```

Add `-delay=<ms>` to any replica to simulate replication/network latency, e.g.
`go run ./replica -config configs/replica1.json -delay=2000`.

## Use the client

```sh
# strong write (needs a majority of 2/3 to ack), then read it back
go run ./client -action=put -node=http://localhost:8081 -key=x -value=10 -consistency=strong
go run ./client -action=get -node=http://localhost:8082 -key=x

# eventual write (returns immediately); read another replica to see staleness
go run ./client -action=put -node=http://localhost:8081 -key=y -value=99 -consistency=eventual
go run ./client -action=get -node=http://localhost:8082 -key=y
```

See [replica/README.md](replica/README.md) and [client/README.md](client/README.md)
for the full API and flag reference.

## Consistency, versioning, conflicts (summary)

- **Eventual**: write is applied locally and replicated asynchronously → high
  availability, temporary staleness.
- **Strong**: write succeeds only after a **majority** of replicas acknowledge
  (`(N/2)+1`, i.e. 2 of 3) → consistency at the cost of availability when a majority
  is unreachable.
- **Versioning**: every write increments the key's `version`.
- **Conflict resolution (Last-Write-Wins)**: higher `version` wins; on equal versions
  the larger `updated_by` (replica id) wins, so all replicas converge deterministically.

## Experiments

The four scenarios required by the assignment are captured in [`results/`](results/):

1. `scenario1.txt` — temporary inconsistency (stale read under eventual consistency).
2. `scenario2.txt` — replica failure: eventual vs strong behaviour (quorum loss).
3. `scenario3.txt` — concurrent conflict resolved by Last-Write-Wins.
4. `scenario4.txt` — effect of replication delay (0 / 500 / 2000 ms) on latency,
   convergence time and stale reads.

A full analysis with the metrics table is in `report.pdf`.
