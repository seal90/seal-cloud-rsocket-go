package main

import (
	"context"
	"log"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
)

func main() {
	// Connect to server
	cli, err := rsocket.Connect().
		Resume().
		Fragment(1024).
		SetupPayload(payload.NewString("Hello", "World")).
		Transport(rsocket.TCPClient().SetHostAndPort("127.0.0.1", 7878).Build()).
		Start(context.Background())
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	// Send request
	result, err := cli.RequestResponse(payload.NewString("你好", "世界")).Block(context.Background())
	if err != nil {
		panic(err)
	}
	log.Println("response:", result)
}
