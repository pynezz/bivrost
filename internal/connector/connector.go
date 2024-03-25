package connector

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/pynezz/bivrost/internal/connector/ipc"
	"github.com/pynezz/bivrost/internal/connector/proto"
	"google.golang.org/grpc"
)

type connectorServer struct {
	proto.UnimplementedConnectorServer
}

type SocketServer struct {
	ipc.UnixSocket
}

func (s *connectorServer) Connect(ctx context.Context, in *proto.ConnectRequest) (*proto.ConnectResponse, error) {
	// Implement your logic to handle the connection here.
	// For now, let's just log the received request and send back a dummy response.
	log.Printf("Received: %v", in)
	return &proto.ConnectResponse{Payload: "Response to " + in.Module}, nil
}

func InitProtobuf(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterConnectorServer(grpcServer, &connectorServer{})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// func NewConnectorServer(port int) {
// 	InitProtobuf(port)
// }

func NewIPC(name string, desc string) (*SocketServer, error) {
	s, err := ipc.NewSocket(name, desc)
	if err != nil {
		return nil, err
	}

	return &SocketServer{*s}, nil
}
