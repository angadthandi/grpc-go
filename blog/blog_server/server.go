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
	"google.golang.org/grpc/reflection"
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

	// Register reflection service on gRPC server.
	reflection.Register(s)

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
		Blog: dataToBlogPb(data),
	}, nil
}

func (s *server) UpdateBlog(
	ctx context.Context,
	req *blogpb.UpdateBlogRequest,
) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("update Blog request")

	blog := req.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
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

	// update internal struct
	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, err = collection.ReplaceOne(
		context.Background(),
		filter,
		data,
	)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update object in mongodb: %v", err),
		)
	}

	return &blogpb.UpdateBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil
}

func (s *server) DeleteBlog(
	ctx context.Context,
	req *blogpb.DeleteBlogRequest,
) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("delete Blog request")

	blogID := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID: %v", err),
		)
	}

	filter := bson.M{"_id": oid}

	ret, err := collection.DeleteOne(
		context.Background(),
		filter,
	)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete blog in mongodb with ID: %v", blogID),
		)
	}

	if ret.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog in mongodb with ID: %v", blogID),
		)
	}

	return &blogpb.DeleteBlogResponse{BlogId: blogID}, nil
}

func (s *server) ListBlog(
	req *blogpb.ListBlogRequest,
	stream blogpb.BlogService_ListBlogServer,
) error {
	fmt.Println("list Blog stream request")

	cur, err := collection.Find(
		context.Background(),
		// nil,
		primitive.D{{}},
	)
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("MongoDB Find error : %v", err),
		)
	}

	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		data := &blogItem{}
		if err := cur.Decode(data); err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Cannot decode blog from cursor error: %v", err),
			)
		}

		stream.Send(&blogpb.ListBlogResponse{Blog: dataToBlogPb(data)})
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("MongoDB internal cursor error: %v", err),
		)
	}

	return nil
}

func dataToBlogPb(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Content:  data.Content,
		Title:    data.Title,
	}
}
