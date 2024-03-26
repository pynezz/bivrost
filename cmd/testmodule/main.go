package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pynezz/bivrost/internal/connector/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// var sock *net.UnixConn

	fmt.Println("Testing module connection...")
	for i, arg := range os.Args {
		switch arg {
		case "proto":
			fmt.Println("Testing gRPC connection...")
			testProtoConnection()
		case "uds":
			fmt.Println("Testing UNIX domain socket connection...")
			testUnixSocketIPC()

		default:
			fmt.Println("Parsing args...")
			fmt.Printf("Arg %d: %s\n", i, arg)
		}
	}
	fmt.Println("End of program. Waiting for SIGINT or SIGTERM...")
	<-c
	fmt.Println("Exiting...")
}

func testProtoConnection() {
	target := "localhost:50051"
	fmt.Printf("Testing gRPC connection to %s...\n", target)

	conn, err := grpc.Dial(
		target, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(),
	)
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

// Connect to a UNIX domain socket
// TODO: Fix
// Connect to a UNIX domain socket
func testUnixSocketIPC() *net.UnixConn {
	path := "/tmp/bivrost/bivrost.sock"
	fmt.Printf("Testing UNIX domain socket connection to %s...\n", path)

	socket, err := net.Dial("unix", path)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	fmt.Println("Connected to UNIX domain socket")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(path)
		os.Exit(1)
	}()

	for {
		buf := make([]byte, 4096)
		copy(buf, []byte("Hello, World!\n"))
		// Accept an incoming connection.
		conn, err := socket.Write(buf)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Sent %d bytes\n", conn)

		// Handle the connection in a separate goroutine.
		go func(conn net.Conn) {
			defer conn.Close()
			// Create a buffer for incoming data.

			// Read data from the connection.
			n, err := socket.Read(buf)
			if err != nil {
				log.Fatal(err)
			}

			// Echo the data back to the connection.
			_, err = conn.Write(buf[:n])
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Received %d bytes\n", n)
			fmt.Printf("Received: %s", buf[:n])

			fmt.Printf("\033[0;32m%s\033[0m\n", "Received: "+string(buf[:n]))

		}(socket)
	}
}
