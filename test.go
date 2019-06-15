package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Message struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func main() {
	//监听退出信号
	closeCh := make(chan os.Signal, 2)
	signal.Notify(closeCh, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGSTOP, syscall.SIGTERM)

	//创建tcp链接
	log.Println("Start listening tcp")
	l, err := net.Listen("tcp", ":8924")

	if err != nil {
		log.Println("fail to listen")
	}

	http.HandleFunc("/query", queryHandler)
	go serveHTTP(l)

	<-closeCh
	log.Println("closing...")
	os.Exit(0)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var msg Message
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	log.Println("query via http")

	_, err = fmt.Fprintf(w, "welcome to my server, %v\n", msg.Name)
	if err != nil {
		log.Printf("write response failed %v", err)
	}
}

func serveHTTP(l net.Listener) {
	if err := http.Serve(l, nil); err != nil {
		log.Printf("while servering http %v\n", err)
	}
}
