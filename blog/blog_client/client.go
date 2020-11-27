package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/angadthandi/grpc-go/blog/blogpb"
	"google.golang.org/grpc"
)

func main() {
	// if we crash go code, we get filename & line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("about to create Blog client connect...")

	opts := grpc.WithInsecure()
	clientConn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer clientConn.Close()

	c := blogpb.NewBlogServiceClient(clientConn)

	fmt.Println("creating blog")
	blog := &blogpb.Blog{
		AuthorId: "Angad",
		Title:    "First Blog",
		Content:  "Blog Content",
	}

	createRet, err := c.CreateBlog(
		context.Background(),
		&blogpb.CreateBlogRequest{Blog: blog},
	)
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
	fmt.Printf("blog created: %v\n", createRet)
	blogID := createRet.GetBlog().GetId()

	// read blog
	_, err = c.ReadBlog(
		context.Background(),
		&blogpb.ReadBlogRequest{BlogId: "5fbfd6c6b2f3e632b058288e"},
	)
	if err != nil {
		fmt.Printf("Error while read: %v\n", err)
	}

	readRet, err := c.ReadBlog(
		context.Background(),
		&blogpb.ReadBlogRequest{BlogId: blogID},
	)
	if err != nil {
		log.Fatalf("unexpected error while read: %v", err)
	}
	fmt.Printf("blog read: %v\n", readRet)

	// update blog
	newBlog := &blogpb.Blog{
		Id:       blogID,
		AuthorId: "Changed Author",
		Title:    "First Blog (edited)",
		Content:  "Blog Content, with some additions!",
	}

	updateRet, err := c.UpdateBlog(
		context.Background(),
		&blogpb.UpdateBlogRequest{
			Blog: newBlog,
		},
	)
	if err != nil {
		log.Fatalf("Error while update: %v", err)
	}
	fmt.Printf("blog updated: %v\n", updateRet)

	// delete blog
	deleteRet, err := c.DeleteBlog(
		context.Background(),
		&blogpb.DeleteBlogRequest{
			BlogId: blogID,
		},
	)
	if err != nil {
		log.Fatalf("Error while delete: %v", err)
	}
	fmt.Printf("blog delete: %v\n", deleteRet)

	// list blog
	retStream, err := c.ListBlog(
		context.Background(),
		&blogpb.ListBlogRequest{},
	)
	if err != nil {
		log.Fatalf("error calling ListBlog RPC: %v", err)
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

		log.Printf("Response from ListBlog: %v\n", msg.GetBlog())
	}
}
