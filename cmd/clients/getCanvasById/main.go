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
		"uid":      "0c07a35c-1962-40ae-947c-1cb6c8cb7602",
		"verified": "true",
	}))
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// Запрос
	req := &canavasv1.GetCanvasByIdRequest{
		CanvasId: "27aae395-13ae-4c35-b725-1ac108c2ae96",
	}

	res, err := client.GetCanvasById(ctx, req)
	if err != nil {
		log.Fatalf("ошибка при вызове GetCanvasByIdRequest: %v", err)
	}

	fmt.Println(res)
}