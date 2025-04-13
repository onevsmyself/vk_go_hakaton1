package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"worker/config"
	"worker/internal/handlers"
	"worker/internal/models"
	"worker/internal/service"
)

var WebPort string
var NumWorkers string

func init() {
	WebPort = *flag.String("web_port", ":80", "Web server port")
	NumWorkers = *flag.String("num", "1", "Number of workers")
	flag.Parse()
}

func main() {
	config.LoadConfig("config/config.yaml")

	filePart := "../worker" + NumWorkers + ".txt"

	channelRequest := make(chan models.SearchRequest, 1000)
	channelResponse := make(chan any, 1000)

	conn, err := net.Dial("tcp", "127.0.0.1:4545")
	if err != nil {
		panic(err)
	}

	fmt.Println(filePart)

	go service.Search(channelRequest, channelResponse, filePart)
	go handlers.Receive(conn, channelRequest)
	go handlers.Send(conn, channelResponse)

	log.Println("Worker started")
	log.Println("Listening on port", WebPort)
	ctx := context.Background()

	<-ctx.Done()
}
