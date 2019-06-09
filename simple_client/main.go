package main

import (
	pb "cmux-demo/pb"
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
)

const (
	address     = "localhost:8924"
	defaultName = "1"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("can not connect: %v", err)
	}

	defer conn.Close()

	c := pb.NewSimpleClient(conn)


	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	} 

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})

	if err != nil {
		fmt.Printf("could not greet: %v", err)
	}
	fmt.Printf("Greeting: %s\n", r.Message)
}
