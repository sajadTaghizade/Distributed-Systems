# Replica — Replicated Key-Value Store node

Each replica is an **independent process** that keeps its own in-memory copy of the
data and talks to its peers over HTTP. Running three of them forms the cluster used
in the experiments.

## Build

From the project root (where `go.mod` lives):

```sh
go build -o bin/replica.exe ./replica
```

You can also run it directly without building: `go run ./replica <flags>`.

## Run

Start three replicas, each in its own terminal. Either pass the topology with flags:

```sh
go run ./replica -id=replica1 -port=8081 -peers=http://localhost:8082,http://localhost:8083
go run ./replica -id=replica2 -port=8082 -peers=http://localhost:8081,http://localhost:8083
go run ./replica -id=replica3 -port=8083 -peers=http://localhost:8081,http://localhost:8082
```

…or load it from a config file (equivalent to the commands above):

```sh
go run ./replica -config configs/replica1.json
go run ./replica -config configs/replica2.json
go run ./replica -config configs/replica3.json
```

### Flags

| Flag       | Default      | Description                                                        |
|------------|--------------|--------------------------------------------------------------------|
| `-id`      | `replica1`   | Replica identifier (also used as the conflict-resolution tiebreak).|
| `-port`    | `8081`       | TCP port to listen on.                                             |
| `-peers`   | _(empty)_    | Comma-separated peer base URLs.                                    |
| `-delay`   | `0`          | Artificial delay (ms) before sending each replication message — simulates network latency. |
| `-config`  | _(empty)_    | Path to a JSON config file; its values override `-id`/`-port`/`-peers`. |

`configs/replicaN.json` format:

```json
{ "id": "replica1", "port": "8081", "peers": ["http://localhost:8082", "http://localhost:8083"] }
```

## HTTP API

### `POST /put` — client write
Body: `{"key":"x","value":"10","consistency":"eventual"}` (`consistency` is `eventual` or `strong`).

- **eventual**: applied locally, replicated to peers asynchronously, returns immediately.
- **strong**: replicated to peers synchronously; succeeds only if a **majority** of the
  cluster (this replica + peers) acknowledges, otherwise returns HTTP 500.

Example responses:
```json
{"status":"success","consistency":"strong","key":"x","version":1,"replicas_updated":3,"quorum":2}
{"status":"error","message":"failed to achieve quorum","replicas_updated":1,"quorum":2}
```

### `GET /get?key=x` — read the **local** copy
```json
{"key":"x","value":"10","version":1,"updated_by":"replica1","served_by":"replica2"}
```
Returns HTTP 404 `{"error":"key not found", ...}` if the key is unknown to this replica.
Because each replica answers from its own copy, a read issued before replication
completes can observe a stale (or missing) value — this is what makes eventual
consistency observable.

### `POST /internal/replicate` — peer-to-peer replication (internal)
Receives `{"key":"x","data":{"value":"10","version":1,"updated_by":"replica1"}}` from
another replica and applies it under the conflict-resolution policy below.

## Versioning & conflict resolution

- **Versioning**: every write to a key increments that key's `version`.
- **Conflict resolution (Last-Write-Wins)**: an incoming value is accepted when the key
  is new **or** its `version` is higher; on **equal** versions the larger `updated_by`
  (replica id) wins. This guarantees a newer value is never overwritten by an older one,
  and that concurrent writes converge to the same winner on every replica.

## Logs

Every replica logs `[PUT]`, `[GET]`, and `[REPLICATE]` events (including whether a
replicated value was applied or rejected), which makes replication and conflict
resolution easy to follow in the experiments.
