package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

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

	// doUnary(c)

	// doServerStreaming(c)

	// doClientStreaming(c)

	doBiDiStreaming(c)
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

func doServerStreaming(
	c greetpb.GreetServiceClient,
) {
	fmt.Println("starting to do Server Streaming RPC...")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Angad",
			LastName:  "Thandi",
		},
	}
	retStream, err := c.GreetManyTimes(
		context.Background(),
		req,
	)
	if err != nil {
		log.Fatalf("error calling GreetManyTimes RPC: %v", err)
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

		log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
	}
}

func doClientStreaming(
	c greetpb.GreetServiceClient,
) {
	fmt.Println("starting to do Client Streaming RPC...")

	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Pammi",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Taran",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Udey",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Angad",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error calling LongGreet RPC: %v", err)
	}

	for _, req := range requests {
		fmt.Printf("Sending req: %v\n", req)

		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	ret, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error receiving response from LongGreet RPC: %v", err)
	}

	fmt.Printf("LongGreet response: %v\n", ret)
}

func doBiDiStreaming(
	c greetpb.GreetServiceClient,
) {
	fmt.Println("starting to do BiDi Streaming RPC...")

	// create stream by invoking client
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("error creating stream for GreetEveryone RPC: %v", err)
		return
	}

	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Pammi",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Taran",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Udey",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Angad",
			},
		},
	}

	waitc := make(chan struct{})
	// send bunch of messages to client (go routine)
	go func() {
		// func to send bunch of messages
		for _, req := range requests {
			fmt.Printf("sending message: %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	// receive bunch of messages from client (go routine)
	go func() {
		// func to receive bunch of messages
		for {
			ret, err := stream.Recv()
			if err == io.EOF {
				// we reached end of stream
				break
			}
			if err != nil {
				log.Fatalf("error receiving: %v", err)
				break
			}
			fmt.Printf("Received: %v\n", ret.GetResult())
		}
		close(waitc)
	}()

	// block until everything is done
	<-waitc
}
