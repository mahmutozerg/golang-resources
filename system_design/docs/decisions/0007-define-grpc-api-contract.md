# **Define gRPC API Contract using Protobuf**

## **Context and Problem Statement**

We are preparing to migrate from a monolithic function-call architecture to a distributed gRPC-based architecture.

Currently, Node and Ring communicate via Go function signatures (Put(string, string, int)).

We need a language-agnostic, strongly-typed interface definition to serve as the contract between the Coordinator (Ring) and the Storage Engine (Node).

## **Decision Drivers**

* **Interoperability:** Future support for clients in other languages.  
* **Performance:** Efficient binary serialization (Proto3) vs Text (JSON).  
* **Binary Safety:** Values can be any binary data (images, structs), not just UTF-8 strings.  
* **Strict Contract:** Enforce input/output structures at compilation time.

## **Decision Outcome**

Chosen option: **"Protobuf with bytes for values"**.

### **API Schema Definition**

We will define a KVStore service with the following RPCs:

1. **Put:**  
   * Input: Key (string), Value (bytes)  
   * Output: Success (bool)  
2. **Get:**  
   * Input: Key (string)  
   * Output: Value (bytes), Found (bool)  
3. **Delete:**  
   * Input: Key (string)  
   * Output: Success (bool)

### **Consequences**

* **Good:** bytes type allows raw binary transport without manual encoding overhead during transport.  
* **Good:** Protobuf generates Go structs automatically, reducing boilerplate.  
* **Neutral:** Storage layer (WAL) must still perform Base64 encoding/decoding because our CSV log format requires text-safe strings. We will convert Proto Bytes \<-\> WAL Base64 String inside the Node.