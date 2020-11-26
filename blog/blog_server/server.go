package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/angadthandi/grpc-go/blog/blogpb"
	// "github.com/mongodb/mongo-go-driver/mongo"
	// "github.com/mongodb/mongo-go-driver/mongo/objectid"
	// "github.com/mongodb/mongo-go-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

var collection *mongo.Collection

type server struct{}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func main() {
	// if we crash go code, we get filename & line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("connecting to mongodb...")
	// connect to mongo client
	mongoDbClient, err := mongo.NewClient(
		options.Client().ApplyURI("mongodb://localhost:27017"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to mongo client: %v", err)
	}
	// ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	// defer cancel()
	err = mongoDbClient.Connect(context.TODO())
	if err != nil {
		log.Fatalf("Failed to connect to mongo client: %v", err)
	}

	collection = mongoDbClient.Database("grpcdb").Collection("blog")

	fmt.Println("about to start blog grpc server...")
	// 50051 is default grpc port
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, &server{})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// wait for Control-C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// block until signal received
	<-ch

	fmt.Println("Stopping server")
	s.Stop()
	fmt.Println("Closing listener")
	lis.Close()
	fmt.Println("Closing MongoDB Connection")
	mongoDbClient.Disconnect(context.TODO())
	fmt.Println("End of main program!")
}
