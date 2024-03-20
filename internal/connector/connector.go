package connector

import (
	"context"
	"log"
	"net"

	"github.com/pynezz/bivrost/internal/connector/proto"
	"google.golang.org/grpc"
)

type connectorServer struct {
	proto.UnimplementedConnectorServer
}

func (s *connectorServer) Connect(ctx context.Context, in *proto.ConnectRequest) (*proto.ConnectResponse, error) {
	// Implement your logic to handle the connection here.
	// For now, let's just log the received request and send back a dummy response.
	log.Printf("Received: %v", in)
	return &proto.ConnectResponse{Payload: "Response to " + in.Module}, nil
}

func Initialize() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterConnectorServer(grpcServer, &connectorServer{})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
