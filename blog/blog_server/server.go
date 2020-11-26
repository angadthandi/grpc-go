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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *server) CreateBlog(
	ctx context.Context,
	req *blogpb.CreateBlogRequest,
) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("create Blog request")
	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	ret, err := collection.InsertOne(
		context.Background(), data,
	)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	oid, ok := ret.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to OID: %v", err),
		)
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func (s *server) ReadBlog(
	ctx context.Context,
	req *blogpb.ReadBlogRequest,
) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("read Blog request")

	blogID := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID: %v", err),
		)
	}

	data := &blogItem{}
	// filter := bson.NewDocument(
	// 	bson.EC.ObjectID("_id", oid),
	// )
	filter := bson.M{"_id": oid}

	ret := collection.FindOne(
		context.Background(),
		filter,
	)

	if err := ret.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with ID: %v", err),
		)
	}

	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Content:  data.Content,
			Title:    data.Title,
		},
	}, nil
}
