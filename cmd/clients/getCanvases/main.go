package main

import (
	"context"
	"fmt"
	"log"
	"time"

	canavasv1 "gitlab.crja72.ru/golang/2025/spring/course/projects/go6/contracts/gen/go/canvas"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := canavasv1.NewCanvasClient(conn)

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		"uid":      "658b2b6b-a2e7-4d10-9afc-43db37c1c91a",
		"verified": "true",
	}))
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// Запрос
	req := &canavasv1.GetCanvasesRequest{
		CanvasIds: []string{"27aae395-13ae-4c35-b725-1ac108c2ae96", "cd6d6ea9-cb7b-4a84-a2f3-102c16a15754"},
	}

	res, err := client.GetCanvases(ctx, req)
	if err != nil {
		log.Fatalf("ошибка при вызове GetCanvasByIdRequest: %v", err)
	}

	fmt.Println(res)
}
