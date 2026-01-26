# 0006-transition-to-distributed-architecture-with-grpc-and-docker

## Context and Problem Statement

Currently, `toy_dynamodb` runs as a monolithic application. The `Ring` (coordinator) and `Node` (storage) structures share the same OS process and memory space. Communication happens via simple function calls.

This architecture prevents us from simulating real-world distributed system challenges such as network latency, timeouts, partial network partitions, and independent process failures. We need to decouple the storage engine from the coordination layer to transform the project from a simulation into a true distributed system.

How should nodes communicate, and how should they be isolated to best represent a production environment?

## Considered Options

* **Transport Layer:**
* **HTTP/1.1 (REST/JSON):** Ubiquitous but verbose and slower due to text serialization.
* **Go `net/rpc` (Gob):** Native and fast, but Go-specific (limits future multi-language interoperability).
* **gRPC (Protobuf):** Industry standard, strongly typed, high performance (HTTP/2), language-agnostic.


* **Isolation & Orchestration:**
* **Localhost Ports:** Run multiple binaries on different ports (e.g., :8001, :8002). Hard to manage IP-based logic.
* **Programmatic Docker (Docker SDK):** The application spawns containers dynamically. High complexity, tight coupling to Docker daemon.
* **Infrastructure-as-Code (Docker Compose):** Define nodes and network via YAML. Application relies on standard DNS/Socket resolution.



## Decision Outcome

Chosen option: **"gRPC with Docker Compose"**, because

* **gRPC** enforces a strict contract (`.proto`) between the Coordinator and Storage nodes, forcing us to handle serialization and network errors explicitly.
* **Protobuf** provides efficient binary serialization, which is crucial for a database system.
* **Docker Compose** allows us to simulate a real network topology where each node has its own IP address and hostname (e.g., `node-1`, `node-2`), decoupling the application logic from infrastructure management.
* It avoids "God Object" anti-patterns where the database application tries to be its own container orchestrator.

### Consequences

* **Good, because** it enables true fault isolation; a crash in one node's container does not kill the coordinator.
* **Good, because** it introduces real-world "Network Jitter" and latency, allowing us to test timeouts and retries realistically.
* **Good, because** we can use Linux `cgroups` (via Docker) to limit memory/CPU for each node, simulating "noisy neighbor" problems later.
* **Bad, because** the development environment becomes more complex (requires rebuilding images and restarting containers for code changes).
* **Bad, because** we must now manage `.proto` file definitions and code generation steps.
* **Bad, because** debugging requires attaching remote debuggers or relying heavily on distributed logging/tracing.

---

### 🗺️ V8 Yol Haritası (Docker & gRPC)

Bu kararı aldıysak, V8 için teknik yapılacaklar listemiz (TODO) şöyle şekillenecek:

1. **Contract (Sözleşme):** `kv.proto` dosyasını oluşturacağız.
* `service KVStore { rpc Put... rpc Get... }`


2. **Server (Node):** `node` paketini, 50051 portunu dinleyen bir `main` uygulamasına (`cmd/server/main.go`) çevireceğiz.
3. **Client (Ring):** `ring` paketindeki `node.Put()` çağrılarını `grpcClient.Put()` ile değiştireceğiz.
4. **Infrastructure:** 3-5 tane node ve 1 tane client ayağa kaldıran bir `docker-compose.yaml` yazacağız.
