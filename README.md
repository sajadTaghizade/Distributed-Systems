# 🌐 Distributed Systems Projects

[![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Python](https://img.shields.io/badge/Python-3.10%2B-3776AB?logo=python&logoColor=white)](https://www.python.org/)
[![Docker](https://img.shields.io/badge/Docker-multi--stage-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](#license)

Welcome to my **Distributed Systems** repository! It contains the computer
assignments from my Distributed Systems coursework at the **University of
Tehran (UT)**, together with a full research project on semantic forwarding
in Named Data Networking.

The assignments progress from raw IPC and RPC between independent processes,
through a multi-VM microservice architecture with a monitoring side-channel,
to a replicated key-value store that trades off consistency for availability
— i.e. the standard distributed-systems ladder: communication → coordination
→ replication. The final project sits on top of that ladder: a gossip-based,
name-based forwarding scheme for NDN/IoT networks, backed by a discrete-event
simulator, 50 unit tests, and an ns-3 cross-validation.

---

## 🛠️ Tech Stack & Tools

| Area | Tools |
|---|---|
| **Languages** | Go (CA1–CA3), Python (Final Project) |
| **Networking** | TCP sockets, JSON-RPC, HTTP/JSON, Named Pipes (FIFOs) |
| **Concurrency** | Goroutines, `sync.WaitGroup`, `GOMAXPROCS` tuning |
| **Distribution patterns** | Client/Server, Pub/Sub, multi-VM service split, quorum replication |
| **Containerization** | Multi-stage Docker builds (`scratch` base images) |
| **Simulation & ML** | Discrete-event simulation, ONNX / MiniLM embeddings, ns-3 (ndnSIM) |
| **Environment** | Linux |

---

## 📂 Repository Layout

```
.
├── CA1/            Low-level IPC → concurrency benchmarking → containerized microservice
├── CA2/            Multi-VM system: web server, auth RPC service, file server, pub/sub monitor
├── CA3/            Replicated key-value store (eventual vs. strong consistency)
└── Final_Project/  Gossip-Scaled Semantic Forwarding in NDN (research project)
```

---

## 📦 Projects Overview

### [CA1 — IPC, Concurrency & Containerized Microservices](./CA1)
Three linked parts moving from OS-level IPC to a deployable service.
* **Part 1 — FIFO-based Client/Worker:** A CLI `Interface` and a background
  `Worker` communicate over two unidirectional Linux named pipes using
  structured JSON requests/responses (`{"operation":"POW","a":2,"b":8}`),
  with strict validation and graceful error propagation.
* **Part 2 — Go Scheduler Benchmarking:** Benchmarks CPU-bound vs. mixed
  (I/O-simulated) workloads across `GOMAXPROCS` and goroutine counts,
  measuring throughput, latency and context-switch overhead to show the gap
  between concurrency and true parallelism.
* **Part 3 — Dockerized HTTP Microservice:** The Part 1 logic is exposed over
  HTTP with a custom logging middleware and shipped as a statically linked,
  zero-dependency binary in a multi-stage `scratch` Docker image.

### [CA2 — Multi-VM Distributed System](./CA2)
A small system split across independent services ("VMs"), each isolated by
responsibility and communicating only over the network:
* **`web-vm`** — the public-facing web server and entry point.
* **`auth-vm`** — a JSON-RPC-over-TCP authentication service with an
  in-memory user store and SHA-256 password hashing; the web tier never
  touches user data directly.
* **`file-vm`** — a dedicated file-serving service.
* **`pubsub`** — an HTTP/JSON publish/subscribe monitor that receives memory
  alerts pushed by the web VM when it exceeds a configured RAM threshold,
  demonstrating an out-of-band monitoring channel decoupled from the request
  path.

### [CA3 — Replicated Key-Value Store](./CA3)
A 3-node replicated key-value store communicating over HTTP, with a CLI
client, built to demonstrate consistency trade-offs first-hand.
* **Eventual consistency:** writes return immediately and replicate
  asynchronously — available, but temporarily stale.
* **Strong consistency:** writes only succeed after a **majority quorum**
  ((N/2)+1 replicas) acknowledges — consistent, but unavailable if the
  majority is unreachable.
* **Conflict resolution:** monotonic per-key versioning with a
  deterministic Last-Write-Wins rule.
* Includes four captured experiments (stale reads, replica failure under
  quorum loss, concurrent write conflicts, and the effect of replication
  delay) and a full write-up in `CA3/report.pdf`.

### [Final Project — Gossip-Scaled Semantic Forwarding in NDN](./Final_Project)
A research project (with M. M. Yari, advised by Dr. M. Shakournia) on
scaling semantic, embedding-based forwarding in Named Data Networking.
Exact-match NDN forwarding fails when a request and a route are phrased
differently; matching by embedding similarity fixes that, but re-running a
transformer per router does not scale as a network grows. This project
shows that **sharing a verified name resolution over anti-entropy gossip**,
instead of caching it locally per router, keeps recognition cost from
growing with the network — encoder calls grow **10%** from 1 to 16 edge
routers versus **60%** for per-router caching alone — and goes on to replace
the usual hand-tuned similarity threshold with a **per-route error budget**
learned online from the network's own Data/Nack feedback, so it needs no
labelled data from the domain it runs on.

* A full discrete-event network simulator (routers, FIB/PIT/Content Store,
  producers, gossip, adversarial and churn models) implemented from scratch.
* Cross-validated against **ns-3 / ndnSIM** (median RTTs agree within 0.11 ms).
* 50 unit tests, a reproducible experiment pipeline, and a detailed,
  self-critical results write-up in [`RESULTS.md`](./Final_Project/RESULTS.md).

---

## 🚀 Getting Started

Each Go project builds and runs independently:

```bash
# CA1 / CA2 / CA3 — from inside the relevant module directory
go build ./...
go run .
```

The Final Project is Python-based and self-contained:

```bash
cd Final_Project
pip install -r requirements.txt
python experiments/fetch_model.py
python experiments/run_experiments.py --all --seeds 20
python -m pytest tests/ -q
```

See each subfolder's own README for exact commands, ports and flags.

---

## 📄 License

This repository is shared for educational purposes. See individual project
folders for any project-specific licensing notes.
