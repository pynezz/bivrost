package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pynezz/bivrost/internal/connector/proto"
	"google.golang.org/grpc"
)

func main() {
	target := "localhost:50051"
	fmt.Printf("Testing gRPC connection to %s...\n", target)
	conn, err := grpc.Dial(target, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := proto.NewConnectorClient(conn)

	secs := 0

	// Example: Send data every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		fmt.Printf("[%d] Trying to connect...\n", secs)
		secs += 5
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		response, err := c.Connect(ctx, &proto.ConnectRequest{
			Module:  "TestModule",
			Method:  "TestMethod",
			Payload: "RandomData",
		})
		if err != nil {
			log.Fatalf("could not connect: %v", err)
		}
		log.Printf("Response: %s", response.GetPayload())
	}
}
