# Client — CLI for the Replicated Key-Value Store

A small command-line client that sends a single `PUT` or `GET` to one replica and
prints the response together with the **measured latency** (used to fill the
PUT/GET latency columns of the report).

## Build

From the project root:

```sh
go build -o bin/client.exe ./client
```

Or run directly: `go run ./client <flags>`.

## Flags

| Flag           | Default                  | Description                                  |
|----------------|--------------------------|----------------------------------------------|
| `-action`      | _(required)_             | `put` or `get`.                              |
| `-node`        | `http://localhost:8081`  | Base URL of the replica to contact.          |
| `-key`         | _(empty)_                | Key to read/write.                           |
| `-value`       | _(empty)_                | Value to write (required for `put`).         |
| `-consistency` | `eventual`               | `eventual` or `strong` (only affects `put`). |

## Examples

```sh
# Strong write (waits for a majority to acknowledge)
go run ./client -action=put -node=http://localhost:8081 -key=x -value=10 -consistency=strong

# Eventual write (returns immediately, replicates in background)
go run ./client -action=put -node=http://localhost:8081 -key=y -value=99 -consistency=eventual

# Read a key from a specific replica
go run ./client -action=get -node=http://localhost:8082 -key=x
```

## Output

```
PUT x=10 [strong] -> http://localhost:8081 | 837.7 ms | HTTP 200
  response: {"consistency":"strong","key":"x","quorum":2,"replicas_updated":3,"status":"success","version":1}

GET x -> http://localhost:8082 | 16.7 ms | HTTP 200
  response: {"key":"x","served_by":"replica2","updated_by":"replica1","value":"10","version":1}
```

Reading the same key from different replicas right after an **eventual** write is the
easiest way to observe temporary inconsistency: the non-coordinator replica may return
`HTTP 404` / an old value until replication completes.
