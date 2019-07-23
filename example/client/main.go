package main

import (
	"context"
	"log"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"github.com/dakalab/micro/example/proto"
)

func main() {
	conn, _ := grpc.Dial(
		"localhost:9999",
		grpc.WithInsecure(),
	)
	var client = proto.NewGreeterClient(conn)

	var ctx, cancel = context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	var req = &proto.HelloRequest{Name: "Hyper"}
	res, err := client.SayHello(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.GetMessage())
}
