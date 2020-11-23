package main

import (
	"context"
	"fmt"
	"log"

	"github.com/angadthandi/grpc-go/greet/greetpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("about to create client connect...")

	clientConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer clientConn.Close()

	c := greetpb.NewGreetServiceClient(clientConn)
	// fmt.Printf("Created client: %f", c)

	doUnary(c)
}

func doUnary(
	c greetpb.GreetServiceClient,
) {
	fmt.Println("starting to do Unary RPC...")

	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Angad",
			LastName:  "Thandi",
		},
	}
	ret, err := c.Greet(
		context.Background(),
		req,
	)
	if err != nil {
		log.Fatalf("error calling Greet RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", ret.Result)
}
