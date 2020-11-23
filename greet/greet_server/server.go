package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/angadthandi/grpc-go/greet/greetpb"
	"google.golang.org/grpc"
)

type server struct{}

func main() {
	fmt.Println("about to start grpc server...")

	// 50051 is default grpc port
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) Greet(
	ctx context.Context,
	req *greetpb.GreetRequest,
) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet func invoked with req: %v\n", req)

	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName
	ret := &greetpb.GreetResponse{
		Result: &greetpb.Greeting{
			FirstName: result,
		},
	}

	return ret, nil
}
