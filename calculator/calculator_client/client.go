package main

import (
	"context"
	"fmt"
	"log"

	"github.com/angadthandi/grpc-go/calculator/calculatorpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("about to create calculator client connect...")

	clientConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer clientConn.Close()

	c := calculatorpb.NewCalculatorServiceClient(clientConn)
	// fmt.Printf("Created client: %f", c)

	doUnary(c)
}

func doUnary(
	c calculatorpb.CalculatorServiceClient,
) {
	fmt.Println("starting to do Sum Unary RPC...")

	req := &calculatorpb.SumRequest{
		FirstNumber:  5,
		SecondNumber: 40,
	}
	ret, err := c.Sum(
		context.Background(),
		req,
	)
	if err != nil {
		log.Fatalf("error calling Calculator RPC: %v", err)
	}
	log.Printf("Response from Sum: %v", ret.SumResult)
}
