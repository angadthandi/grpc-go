package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/angadthandi/grpc-go/calculator/calculatorpb"
	"google.golang.org/grpc"
)

type server struct{}

func main() {
	fmt.Println("starting calculator server...")

	// 50051 is default grpc port
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) Sum(
	ctx context.Context,
	req *calculatorpb.SumRequest,
) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Sum func invoked with req: %v\n", req)

	num1 := req.FirstNumber
	num2 := req.SecondNumber
	sum := num1 + num2
	ret := &calculatorpb.SumResponse{
		SumResult: sum,
	}

	return ret, nil
}
