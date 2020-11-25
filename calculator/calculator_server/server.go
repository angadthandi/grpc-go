package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

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

func (s *server) PrimeNumberDecomposition(
	req *calculatorpb.PrimeRequest,
	stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer,
) error {
	fmt.Printf("PrimeNumberDecomposition func invoked with req: %v\n", req)

	k := int64(2)
	N := req.GetNumber()
	for N > 1 {
		if N%k == 0 { // if k evenly divides into N
			// print k      // this is a factor
			ret := &calculatorpb.PrimeResponse{
				PrimeResult: k,
			}

			stream.Send(ret)
			time.Sleep(1000 * time.Millisecond)

			N = N / k // divide N by k so that we have the rest of the number left.
		} else {
			k = k + 1
		}
	}

	return nil
}

func (s *server) ComputeAverage(
	stream calculatorpb.CalculatorService_ComputeAverageServer,
) error {
	fmt.Println("ComputeAverage func invoked with streaming req")

	sum := int64(0)
	cnt := int64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			average := float32(sum) / float32(cnt)
			fmt.Printf(
				"Computing Average response from sum: %v; count: %v; average: %v\n",
				sum, cnt, average,
			)

			ret := &calculatorpb.ComputeAverageResponse{
				AverageResult: average,
			}

			return stream.SendAndClose(ret)
		}
		if err != nil {
			log.Fatalf("error reading client stream: %v", err)
		}

		sum += req.GetNumber()
		cnt += 1
	}
}

func (s *server) FindMaximum(
	stream calculatorpb.CalculatorService_FindMaximumServer,
) error {
	fmt.Println("FindMaximum func invoked with bidi-streaming req")

	max := int64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("error reading client stream: %v", err)
			return err
		}

		num := req.GetNumber()
		if num > max {
			err = stream.Send(&calculatorpb.FindMaximumResponse{
				MaxResult: num,
			})
			if err != nil {
				log.Fatalf("error sending response stream: %v", err)
				return err
			}

			max = num
		}
	}
}
