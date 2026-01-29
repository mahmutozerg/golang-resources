# Transiiton to Distributed Architecture with Grpc and Docker

## Context and Problem Statement

Currently, `toy_dynamodb` runs as a monolithic application. The `Ring` (coordinator) and `Node` (storage) structures share the same OS process and memory space. Communication happens via simple function calls.

This architecture prevents us from simulating real-world distributed system challenges such as network latency, timeouts, partial network partitions, and independent process failures. We need to decouple the storage engine from the coordination layer to transform the project from a simulation into a true distributed system.

How should nodes communicate, and how should they be isolated to best represent a production environment?

## Considered Options

- **Transport Layer:**
  - **HTTP/1.1 (REST/JSON):** Ubiquitous but verbose and slower due to text serialization.
  - **Go `net/rpc` (Gob):** Native and fast, but Go-specific (limits future multi-language interoperability).
  - **gRPC (Protobuf):** Industry standard, strongly typed, high performance (HTTP/2), language-agnostic.

- **Isolation & Orchestration:**
  - **Localhost Ports:** Run multiple binaries on different ports (e.g., :8001, :8002). Hard to manage IP-based logic.
  - **Programmatic Docker (Docker SDK):** The application spawns containers dynamically. High complexity, tight coupling to Docker daemon.
  - **Infrastructure-as-Code (Docker Compose):** Define nodes and network via YAML. Application relies on standard DNS/Socket resolution.

## Decision Outcome

Chosen option: **"gRPC with Docker Compose"**, because

- **gRPC** enforces a strict contract (`.proto`) between the Coordinator and Storage nodes, forcing us to handle serialization and network errors explicitly.
- **Protobuf** provides efficient binary serialization, which is crucial for a database system.
- **Docker Compose** allows us to simulate a real network topology where each node has its own IP address and hostname (e.g., `node-1`, `node-2`), decoupling the application logic from infrastructure management.
- It avoids "God Object" anti-patterns where the database application tries to be its own container orchestrator.

### Implementation Details

To replicate the distributed environment locally, we use `docker-compose.yaml`. Since we do not use a dynamic service discovery tool (like Consul or Etcd) yet, the cluster topology is statically defined.

- **Explicit Service Definition:** Each storage node must be explicitly defined as a separate service in the compose file (e.g., `node-1`, `node-2`, `node-3`).
- **Networking:** We utilize a custom bridge network. This allows nodes to communicate and enables the client to resolve nodes via their service names or mapped localhost ports.
- **Persistence Mapping:** Docker Volumes must be mapped to the host's `./wal` directory (e.g., `./wal/node-1:/root/wal`) to ensure data survives container restarts.

**Operational Workflow:**

1.  Define nodes in `docker-compose.yml`.
2.  Run `docker-compose up --build` to start the cluster.
3.  The Client (Ring) connects to the exposed ports (e.g., `:50051`, `:50052`).

### Consequences

- **Good, because** it enables true fault isolation; a crash in one node's container does not kill the coordinator.
- **Good, because** it introduces real-world "Network Jitter" and latency, allowing us to test timeouts and retries realistically.
- **Good, because** we can use Linux `cgroups` (via Docker) to limit memory/CPU for each node, simulating "noisy neighbor" problems later.
- **Bad, because** the development environment becomes more complex (requires rebuilding images and restarting containers for code changes).
- **Bad, because** we must now manage `.proto` file definitions and code generation steps.
- **Bad, because** debugging requires attaching remote debuggers or relying heavily on distributed logging/tracing.
