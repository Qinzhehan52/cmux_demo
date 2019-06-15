package handler

import (
	"cmux_demo/pb"
	"context"
	"log"
)

type QueryHandler struct{}

func (s *QueryHandler) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReplay, error) {
	log.Printf("[sayHello]req.Name=%v", req.Name)
	return &pb.HelloReplay{Message: "hello======>" + req.Name + "\n"}, nil
}
