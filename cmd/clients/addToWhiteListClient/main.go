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
		"uid":      "c726593c-b6d6-4bbf-9a01-dd379f62082f",
		"verified": "true",
	}))
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// Запрос
	req := &canavasv1.AddToWhiteListRequest{
		CanvasId: "cd6d6ea9-cb7b-4a84-a2f3-102c16a15754",
		MemberId: "0c07a36c-1962-40ae-947c-1cb6c8cb7602",
	}

	res, err := client.AddToWhiteList(ctx, req)
	if err != nil {
		log.Fatalf("ошибка при вызове AddToWhiteList: %v", err)
	}

	fmt.Printf("Успешно добавлен в whitelist, canvasID: %s\n", res.GetCanvasId())
}
