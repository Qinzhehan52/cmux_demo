package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"cmux_demo/pb"

	"github.com/soheilhy/cmux"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
)

type httpServer struct{}

func (*httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Querying via http")
	b, err := json.Marshal(r.Header)

	if err != nil {
		log.Println("resolve http header failed")
		return
	}

	log.Printf("headers=%s", string(b))
	w.Header().Set("content-type", "application/json")
	fmt.Fprint(w, string(b))
}

func echoServer(ws *websocket.Conn) {
	log.Println("req via ws")
	if _, err := io.Copy(ws, ws); err != nil {
		panic(err)
	}
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
	s := &http.Server{
		Handler: &httpServer{},
	}
	if err := s.Serve(l); err != cmux.ErrListenerClosed {
		panic(err)
	}
}

func serveWS(l net.Listener) {
	s := &http.Server{
		Handler: websocket.Handler(echoServer),
	}
	if err := s.Serve(l); err != nil {
		log.Printf("while serving websocket %v\n", err)
	}
}

func main() {
	//监听退出信号
	closeCh := make(chan os.Signal, 1)
	signal.Notify(closeCh, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGSTOP, syscall.SIGTERM)

	//创建tcp链接
	log.Println("Start listening tcp")
	l, err := net.Listen("tcp", ":8924")

	if err != nil {
		log.Println("fail to listen")
	}

	//cmux开始
	tcpm := cmux.New(l)
	grpc1 := tcpm.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	wsl := tcpm.Match(cmux.HTTP1HeaderField("Upgrade", "websocket"))
	http1 := tcpm.Match(cmux.HTTP1Fast())
	http2 := tcpm.Match(cmux.HTTP2())

	//开始服务
	go serveGPRC(grpc1)
	go serveWS(wsl)
	go serveHTTP(http1)
	go serveHTTP(http2)

	if err := tcpm.Serve(); err != nil && !strings.Contains(err.Error(),
		"use of closed network connection") {
		log.Printf("while servering cmux %v\n", err)
	}

	sig := <-closeCh
	log.Printf("receive signal:%v\n", sig)
	os.Exit(0)
}
