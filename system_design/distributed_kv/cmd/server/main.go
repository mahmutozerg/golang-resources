package main

import (
	"context"
	"log"
	"net"
	"os"
	"toy_dynamodb/pkg/node"
	kv "toy_dynamodb/proto"

	"google.golang.org/grpc"
)

type server struct {
	kv.UnimplementedKVStoreServer
	node *node.Node
}

func (s *server) Get(ctx context.Context, r *kv.GetRequest) (*kv.GetResponse, error) {

	res, success := s.node.Get(r.Key)

	return &kv.GetResponse{
		Value: []byte(res),
		Found: success,
	}, nil
}

func (s *server) Put(ctx context.Context, r *kv.PutRequest) (*kv.PutResponse, error) {

	err := s.node.Put(r.Key, string(r.Value))

	if err != nil {
		return &kv.PutResponse{
			Success: false,
		}, err
	}

	return &kv.PutResponse{
		Success: true,
	}, nil

}

func (s *server) Delete(ctx context.Context, r *kv.DeleteRequest) (*kv.DeleteResponse, error) {
	err := s.node.Del(r.Key)

	if err != nil {
		return &kv.DeleteResponse{
			Success: false,
		}, err
	}

	return &kv.DeleteResponse{
		Success: true,
	}, nil
}

func main() {

	nn := os.Getenv("NODE_NAME")
	if nn == "" {
		log.Fatal("NODE_NAME variable must be set before creating grpc server")
	}
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to create tcp listener %v", err)
	}

	defer listener.Close()

	n, err := node.New(nn)

	if err != nil {
		log.Fatalf("%v failed to create node", err)
	}
	grpcServer := grpc.NewServer()
	kv.RegisterKVStoreServer(grpcServer, &server{node: n})

	grpcServer.Serve(listener)
}
