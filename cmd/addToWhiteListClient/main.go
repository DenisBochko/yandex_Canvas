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
		"uid":      "a1b2c3d4-e5f6-7890-abcd-1234567890ef",
		"verified": "true",
	}))
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// Запрос
	req := &canavasv1.AddToWhiteListRequest{
		CanvasId: "f6cfb75f-f00c-46a7-b061-01b853f1bae9",
		MemberId: "cae633e1-e51e-4d87-9d51-2181e992a255",
	}

	res, err := client.AddToWhiteList(ctx, req)
	if err != nil {
		log.Fatalf("ошибка при вызове AddToWhiteList: %v", err)
	}

	fmt.Printf("Успешно добавлен в whitelist, canvasID: %s\n", res.GetCanvasId())
}
