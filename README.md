# Redigo

A Redis-like in-memory key-value store written in Go with zero external dependencies.

## Features

- **HTTP API** — RESTful endpoints for GET, PUT, DELETE, HEAD, and range SCAN
- **In-Memory Storage** — `map[string][]byte` with defensive copying and `sync.RWMutex`
- **B-Tree Index** — Standalone B-tree with ordered range scan, insert, search, delete
- **Write-Ahead Log (WAL)** — Binary WAL for persistence with replay on restart
- **WAL Compaction** — Atomic snapshot + rename to shrink WAL files
- **TTL (Time-To-Live)** — Per-key expiration with background sweep goroutine
- **Graceful Shutdown** — SIGINT/SIGTERM handling with clean goroutine lifecycle

## Architecture

```
HTTP Handler
    │
    ▼
TTLStore         ← expires map + background sweep goroutine
    │
    ▼
WalStore         ← binary WAL file + sync on every write
    │
    ▼
MemoryStore      ← map[string][]byte + RWMutex
```

Each layer implements the same `store.Store` interface. Fully composable via dependency injection.

## Quick Start

```bash
# build and run
go build -o redigo ./cmd/redigo
./redigo -addr :8080 -wal data.wal -sweep 1s

# or
go run ./cmd/redigo
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8080` | HTTP listen address |
| `-wal` | `data.wal` | WAL file path |
| `-sweep` | `1s` | TTL sweep interval |

## HTTP API

### Set a key

```bash
curl -X PUT http://localhost:8080/keys/mykey -d "hello"
```

### Set with TTL

```bash
curl -X PUT "http://localhost:8080/keys/temp?ttl=30s" -d "short-lived"
```

### Get a key

```bash
curl http://localhost:8080/keys/mykey
```

### Check existence (HEAD)

```bash
curl -I http://localhost:8080/keys/mykey
# 200 = exists, 404 = not found
```

### Delete a key

```bash
curl -X DELETE http://localhost:8080/keys/mykey
```

### Range scan (ordered keys)

```bash
curl "http://localhost:8080/keys?start=apple&end=cherry"
# returns JSON array of {Key, Value} pairs in sorted order
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Empty key / bad request / missing scan params |
| 404 | Key not found |
| 405 | Wrong HTTP method |

## B-Tree

A standalone B-tree implementation in `store/`. Supports:

- **Insert** with automatic node splitting
- **Search** via binary search within nodes
- **Delete** with predecessor/successor replacement and merge
- **Range Scan** via in-order traversal
- **Update** — inserting an existing key updates its value

### Usage as a drop-in store

```go
import "github.com/abeer-srivastava/redigo/store"

// swap MemoryStore for BTreeStore in one line
tree := store.NewBTreeStore(2)

// works with WAL, TTL, HTTP — no changes needed above
wal, _ := persistence.NewWalStore("data.wal", tree)
kvStore := ttl.NewTTLStore(wal, time.Second)
srv := server.NewServer(":8080", kvStore)
```

### Usage as a standalone data structure

```go
b := store.NewBtree(3) // minimum degree 3
b.Insert("key1", []byte("value1"))
val, ok := b.Search("key1")
b.Delete("key1")
results := b.Scan("key1", "key9") // ordered range scan
```

## WAL Compaction

Over time the WAL accumulates duplicate entries (same key written many times). Compaction rewrites the WAL with only the latest values:

```go
// compact the WAL — shrinks file, preserves all live data
err := walStore.Compact()
```

Compaction is atomic — uses `os.Rename` which is atomic on Linux/macOS. Either the old file or the new file exists, never a partial state.

## Testing

```bash
# run all tests
go test ./...

# with verbose output
go test ./... -v

# with race detector
go test ./... -race

# benchmarks
go test ./store/ -bench=BenchmarkBtree -benchtime=1s
```

## Project Structure

```
redigo/
├── cmd/redigo/main.go          — entry point, signal handling
├── server/
│   ├── server.go               — HTTP server, route registration
│   └── handler.go              — HTTP handlers (CRUD + scan)
├── store/
│   ├── store.go                — Store interface + sentinel errors
│   ├── memory.go               — map-backed in-memory store
│   ├── btree.go                — B-tree struct + Search/Insert/Delete/Scan
│   ├── btree_node.go           — B-tree node operations
│   └── btree_store.go          — BTreeStore implementing Store interface
├── persistence/
│   ├── wal.go                  — Write-Ahead Log with replay
│   └── compaction.go           — WAL compaction (atomic snapshot)
├── ttl/
│   └── ttl.go                  — TTL decorator with background sweep
└── go.mod
```

## Module

```
github.com/abeer-srivastava/redigo
```

Go 1.26.1 · Zero external dependencies
