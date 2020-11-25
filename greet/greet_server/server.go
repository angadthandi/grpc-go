package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

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

func (s *server) GreetManyTimes(
	req *greetpb.GreetManyTimesRequest,
	stream greetpb.GreetService_GreetManyTimesServer,
) error {
	fmt.Printf("GreetManyTimes func invoked with req: %v\n", req)

	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " number " + strconv.Itoa(i)
		ret := &greetpb.GreetManyTimesResponse{
			Result: &greetpb.Greeting{
				FirstName: result,
			},
		}

		stream.Send(ret)
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (s *server) LongGreet(
	stream greetpb.GreetService_LongGreetServer,
) error {
	fmt.Println("LongGreet func invoked with streaming req")

	result := ""
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// we reached end of stream
			ret := &greetpb.LongGreetResponse{
				Result: result,
			}

			return stream.SendAndClose(ret)
		}
		if err != nil {
			log.Fatalf("error reading client stream: %v", err)
		}

		result += "Hello " + req.GetGreeting().GetFirstName() + "!"
	}
}

func (s *server) GreetEveryone(
	stream greetpb.GreetService_GreetEveryoneServer,
) error {
	fmt.Println("GreetEveryone func invoked with bidi-streaming req")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// we reached end of stream
			return nil
		}
		if err != nil {
			log.Fatalf("error reading client stream: %v", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName + "!"

		err = stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if err != nil {
			log.Fatalf("error sending data to client: %v", err)
			return err
		}
	}
}
