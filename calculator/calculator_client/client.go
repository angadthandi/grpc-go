package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/angadthandi/grpc-go/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	doServerStreaming(c)

	doClientStreaming(c)

	doBiDiStreaming(c)

	doErrorUnary(c)
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

func doServerStreaming(
	c calculatorpb.CalculatorServiceClient,
) {
	fmt.Println("starting to do Server Streaming RPC...")

	req := &calculatorpb.PrimeRequest{
		Number: 120,
	}
	retStream, err := c.PrimeNumberDecomposition(
		context.Background(),
		req,
	)
	if err != nil {
		log.Fatalf("error calling PrimeNumberDecomposition RPC: %v", err)
	}

	for {
		msg, err := retStream.Recv()
		if err == io.EOF {
			// we reached end of stream
			break
		}
		if err != nil {
			log.Fatalf("error reading stream: %v", err)
		}

		log.Printf("Response from PrimeNumberDecomposition: %v", msg.GetPrimeResult())
	}
}

func doClientStreaming(
	c calculatorpb.CalculatorServiceClient,
) {
	fmt.Println("starting to do Client Streaming RPC...")

	requests := []*calculatorpb.ComputeAverageRequest{
		&calculatorpb.ComputeAverageRequest{
			Number: 1,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 2,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 3,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 4,
		},
	}

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("error calling ComputeAverage RPC: %v", err)
	}

	for _, req := range requests {
		fmt.Printf("Sending req: %v\n", req)

		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	ret, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error receiving ComputeAverage response RPC: %v", err)
	}

	fmt.Printf("ComputeAverage response: %v\n", ret.GetAverageResult())
}

func doBiDiStreaming(
	c calculatorpb.CalculatorServiceClient,
) {
	fmt.Println("starting to do BiDi Streaming RPC...")

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("error creating stream for FindMaximum: %v", err)
		return
	}

	requests := []*calculatorpb.FindMaximumRequest{
		&calculatorpb.FindMaximumRequest{
			Number: 1,
		},
		&calculatorpb.FindMaximumRequest{
			Number: 5,
		},
		&calculatorpb.FindMaximumRequest{
			Number: 3,
		},
		&calculatorpb.FindMaximumRequest{
			Number: 6,
		},
		&calculatorpb.FindMaximumRequest{
			Number: 2,
		},
		&calculatorpb.FindMaximumRequest{
			Number: 20,
		},
	}

	waitc := make(chan struct{})

	// send
	go func() {
		for _, req := range requests {
			fmt.Printf("sending request: %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	// receive
	go func() {
		for {
			ret, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error receiving: %v", err)
				break
			}

			fmt.Printf("received : %v\n", ret.GetMaxResult())
		}
		close(waitc)
	}()

	<-waitc
}

func doErrorUnary(
	c calculatorpb.CalculatorServiceClient,
) {
	fmt.Println("starting to do Error Unary RPC...")

	// correct call
	doErrorCall(c, 10)

	// error call
	doErrorCall(c, -10)
}

func doErrorCall(
	c calculatorpb.CalculatorServiceClient,
	num int,
) {
	n := int64(num)
	req := &calculatorpb.SquareRootRequest{
		Number: n,
	}
	ret, err := c.SquareRoot(
		context.Background(),
		req,
	)
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// grpc error
			fmt.Printf("Error message from grpc: %v\n", respErr.Message())
			fmt.Printf("Error code from grpc: %v\n", respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("Negative number sent!")
				return
			}
		} else {
			log.Fatalf("Error calling SquareRoot: %v\n", err)
			return
		}
	}
	log.Printf("Response SquareRoot of %v is %v\n", n, ret.GetNumberRoot())
}
