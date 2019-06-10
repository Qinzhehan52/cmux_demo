package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"

	"cmux_demo/pb"
)

func queryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Querying via http")
	w.WriteHeader(200)
}

type grpcServer struct{}

func (s *grpcServer) SayHello(ctx context.Context, 
	req *pb.HelloRequest) (*pb.HelloReplay, error) {
	log.Printf("[sayHello]req.Name=%v", req.Name)

	return &pb.HelloReplay{Message: "hello======>" + req.Name + "\n"}, nil
}

func serveGPRC(l net.Listener) {
	s := grpc.NewServer()
	pb.RegisterSimpleServer(s, &grpcServer{})
	if err := s.Serve(l); err != nil {
		log.Printf("while servering grpc %v\n", err)
	}

}

func serveHTTP(l net.Listener) {
	if err := http.Serve(l, nil); err != nil {
		log.Printf("while servering http %v\n", err)
	}
}

func main() {
	//监听退出信号
	closeCh := make(chan os.Signal, 2)
	signal.Notify(closeCh, os.Interrupt, syscall.SIGTERM)

	//创建tcp链接
	log.Println("Start listening tcp")
	l, err := net.Listen("tcp", ":8924")

	if err != nil {
		log.Println("fail to listen")
	}

	//cmux开始
	tcpm := cmux.New(l)
	http1 := tcpm.Match(cmux.HTTP1Fast())
	grpc1 := tcpm.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	http2 := tcpm.Match(cmux.HTTP2())

	http.HandleFunc("/query", queryHandler)

	//开始服务
	go serveGPRC(grpc1)
	go serveHTTP(http1)
	go serveHTTP(http2)

	go func() {
		<-closeCh
		l.Close()
	}()

	if err := tcpm.Serve(); err != nil && !strings.Contains(err.Error(),
		"use of closed network connection") {
		log.Printf("while servering cmux %v\n", err)
	}
}
