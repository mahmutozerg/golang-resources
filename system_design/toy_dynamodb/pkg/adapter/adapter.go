package adapter

import (
	"context"
	"toy_dynamodb/pkg/node"
	kv "toy_dynamodb/proto"

	"google.golang.org/grpc"
)

type LocalClient struct {
	node *node.Node
}

func NewLocalClient(n *node.Node) *LocalClient {
	return &LocalClient{node: n}
}

func (l *LocalClient) Put(ctx context.Context, in *kv.PutRequest, opts ...grpc.CallOption) (*kv.PutResponse, error) {
	err := l.node.Put(in.Key, string(in.Value))
	if err != nil {
		return &kv.PutResponse{Success: false}, err
	}
	return &kv.PutResponse{Success: true}, nil
}

func (l *LocalClient) Get(ctx context.Context, in *kv.GetRequest, opts ...grpc.CallOption) (*kv.GetResponse, error) {
	val, found := l.node.Get(in.Key)

	return &kv.GetResponse{
		Value: []byte(val),
		Found: found,
	}, nil
}

func (l *LocalClient) Delete(ctx context.Context, in *kv.DeleteRequest, opts ...grpc.CallOption) (*kv.DeleteResponse, error) {
	err := l.node.Del(in.Key)
	if err != nil {
		return &kv.DeleteResponse{Success: false}, err
	}
	return &kv.DeleteResponse{Success: true}, nil
}
