package canvas

import (
	"context"
	"errors"
	"time"

	canavasv1 "github.com/DenisBochko/yandex_contracts/gen/go/canvas"
	"google.golang.org/grpc"
)

// TOOD: GetWhiteList(context.Context, *GetWhiteListRequest) (*GetWhiteListResponse, error)

type CanvasServer struct {
	canavasv1.CanvasServer
}

func Register(gRPC *grpc.Server) {
	canavasv1.RegisterCanvasServer(gRPC, &CanvasServer{})
}

func (c *CanvasServer) CreateCanvas(ctx context.Context, req *canavasv1.CreateCanvasRequest) (*canavasv1.CreateCanvasResponse, error) {
	return &canavasv1.CreateCanvasResponse{
		CanvasId: "1234567890",
	}, nil
}

func (c *CanvasServer) GetCanvasById(ctx context.Context, req *canavasv1.GetCanvasByIdRequest) (*canavasv1.GetCanvasByIdResponse, error) {
	time.Sleep(time.Second * 8)
	return nil, errors.New("not implemented")
}

func (c *CanvasServer) GetCanvases(ctx context.Context, req *canavasv1.GetCanvasesRequest) (*canavasv1.GetCanvasesResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *CanvasServer) UploadImage(ctx context.Context, req *canavasv1.UploadImageRequest) (*canavasv1.UploadImageResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *CanvasServer) JoinToCanvas(ctx context.Context, req *canavasv1.JoinToCanvasRequest) (*canavasv1.JoinToCanvasResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *CanvasServer) AddToWhiteList(ctx context.Context, req *canavasv1.AddToWhiteListRequest) (*canavasv1.AddToWhiteListResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *CanvasServer) UpdateCanvas(ctx context.Context, req *canavasv1.UpdateCanvasRequest) (*canavasv1.UpdateCanvasResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *CanvasServer) DeleteCanvas(ctx context.Context, req *canavasv1.DeleteCanvasRequest) (*canavasv1.DeleteCanvasResponse, error) {
	return nil, errors.New("not implemented")
}

// func (s *Server) GetWhiteList(ctx context.Context, req *canavasv1.GetWhiteListRequest) (*canavasv1.GetWhiteListResponse, error) {
// 	return nil, errors.New("not implemented")
// }
