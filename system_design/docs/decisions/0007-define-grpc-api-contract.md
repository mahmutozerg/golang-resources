# Define gRPC API Contract using Protobuf

## Context and Problem Statement

We are preparing to migrate from a monolithic function-call architecture to a distributed gRPC-based architecture.

Currently, Node and Ring communicate via Go function signatures (`Put(string, string, int)`).

We need a language-agnostic, strongly-typed interface definition to serve as the contract between the Coordinator (Ring) and the Storage Engine (Node). Furthermore, we need a strategy to integrate this contract without tightly coupling the core logic to the network layer or sacrificing testability.

## Decision Drivers

- **Interoperability:** Future support for clients in other languages.
- **Performance:** Efficient binary serialization (Proto3) vs Text (JSON).
- **Binary Safety:** Values can be any binary data (images, structs), not just UTF-8 strings.
- **Strict Contract:** Enforce input/output structures at compilation time.
- **Testability:** The ability to test core logic without spinning up full network infrastructure.

## Decision Outcome

Chosen option: **"Protobuf with bytes for values"**.

### API Schema Definition

We will define a `KVStore` service with the following RPCs:

1. **Put:**
   - Input: `Key` (string), `Value` (bytes)
   - Output: `Success` (bool)
2. **Get:**
   - Input: `Key` (string)
   - Output: `Value` (bytes), `Found` (bool)
3. **Delete:**
   - Input: `Key` (string)
   - Output: `Success` (bool)

### Implementation Patterns

To ensure clean separation of concerns and maintain fast testing cycles, we will adopt the following patterns:

#### 1. Server-Side Adapter (Ports and Adapters)

We will **not** modify the core `Node` logic to be aware of gRPC. Instead, we will use the **Adapter Pattern**:

- A new `GRPCServer` struct will wrap the `Node` struct.
- It will implement the generated `KVStoreServer` interface.
- It handles the translation between Protobuf types (e.g., `[]byte`) and domain types (e.g., `string`, Base64 encoding).

#### 2. Local In-Memory Testing Strategy

Since the Ring coordinator depends on the `KVStoreClient` **interface** (not a concrete network client), we will enable running integration tests without Docker.

We can define a `LocalClient` struct that satisfies the protobuf interface but calls the `Node` directly:

```go
// Example Implementation for Local Testing
type LocalClient struct {
    node *node.Node
}

// Implement the gRPC interface methods directly calling the node
func (l *LocalClient) Put(ctx context.Context, in *kv.PutRequest, opts ...grpc.CallOption) (*kv.PutResponse, error) {
    err := l.node.Put(in.Key, string(in.Value)) // Type conversion: bytes -> string
    if err != nil {
        return &kv.PutResponse{Success: false}, err
    }
    return &kv.PutResponse{Success: true}, nil
}

// Usage in Test Files (No Docker required):
func TestRingLogic() {
    // 1. Create the engine directly (bypass gRPC server)
    storageEngine, _ := node.New("local-memory-node")

    // 2. Wrap it with the adapter
    adapter := &LocalClient{node: storageEngine}

    // 3. Inject into Ring (bypass grpc.Dial)
    myRing.RegisterClient("node-1", adapter)
}
```

### Consequences

- **Good:** `bytes` type allows raw binary transport without manual encoding overhead during transport.
- **Good:** Protobuf generates Go structs automatically, reducing boilerplate.
- **Good:** The Adapter pattern keeps the core storage engine isolated from transport details (HTTP/gRPC agnostic).
- **Good:** In-memory testing strategy prevents the development lifecycle from becoming slow due to Docker/Network overhead.
- **Neutral:** Storage layer (WAL) must still perform Base64 encoding/decoding because our CSV log format requires text-safe strings. We will convert Proto Bytes <-> WAL Base64 String inside the Adapter.

```

```

```

```
