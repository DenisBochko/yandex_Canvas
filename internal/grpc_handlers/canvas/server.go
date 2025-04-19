package canvas

import (
	"context"
	"slices"

	"github.com/DenisBochko/yandex_Canvas/internal/domain/models"
	canavasv1 "github.com/DenisBochko/yandex_contracts/gen/go/canvas"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TOOD: GetWhiteList
// TODO: Сделать человеческую обработку ошибок

type CanvasService interface {
	CreateCanvas(ctx context.Context, name string, width int32, height int32, ownerID string, privacy string) (string, error) // возвращает id созданного канваса
	GetCanvasById(ctx context.Context, id string) (models.Canvas, error)
	GetCanvases(ctx context.Context, canvasIDs []string) ([]models.Canvas, error)
	UploadImage(ctx context.Context, canvasID string, image []byte) (string, error) // возвращает id загруженного изображения

	// Одобрение по почте (отправляем owner`у на почту)
	JoinToCanvas(ctx context.Context, canvasID string, userID string) (string, error) // возвращает id канваса

	AddToWhiteList(ctx context.Context, canvasID string, userID string) (string, error)             // возвращает id канваса
	UpdateCanvas(ctx context.Context, canvasID string, name string, privacy string) (string, error) // возвращает id обновлённого канваса
	DeleteCanvas(ctx context.Context, canvasID string) (string, error)                              // возвращает id удалённого канваса
	// GetWhiteList(ctx context.Context, canvasID string) ([]string, error)                         // возвращает id пользователей, которые могут редактировать канвас
}

const (
	CanvasPrivacyPublic  = "public"  // публичный холст
	CanvasPrivacyPrivate = "private" // приватный холст
	CanvasPrivacyFriends = "friends" // холст доступен только друзьям
)

type CanvasServer struct {
	canavasv1.CanvasServer
	canvasService CanvasService
}

func Register(gRPC *grpc.Server, canvasService CanvasService) {
	canavasv1.RegisterCanvasServer(gRPC, &CanvasServer{canvasService: canvasService})
}

func (c *CanvasServer) CreateCanvas(ctx context.Context, req *canavasv1.CreateCanvasRequest) (*canavasv1.CreateCanvasResponse, error) {
	// базовая проверка
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetOwnerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ownerID is required")
	}
	if !slices.Contains([]string{CanvasPrivacyPublic, CanvasPrivacyPrivate, CanvasPrivacyFriends}, req.Privacy) {
		return nil, status.Error(codes.InvalidArgument, "canvas privacy is not included (public, private, friends)")
	}

	canvasID, err := c.canvasService.CreateCanvas(
		ctx,
		req.GetName(),
		req.GetWidth(),
		req.GetHeight(),
		req.GetOwnerId(),
		req.GetPrivacy(),
	)

	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &canavasv1.CreateCanvasResponse{
		CanvasId: canvasID,
	}, nil
}

func (c *CanvasServer) GetCanvasById(ctx context.Context, req *canavasv1.GetCanvasByIdRequest) (*canavasv1.GetCanvasByIdResponse, error) {
	if req.GetCanvasId() == "" {
		return nil, status.Error(codes.InvalidArgument, "canvasID is required")
	}

	canvas, err := c.canvasService.GetCanvasById(ctx, req.GetCanvasId())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &canavasv1.GetCanvasByIdResponse{
		Canvas: &canavasv1.Canvas{
			CanvasId:   canvas.ID,
			Name:       canvas.Name,
			Width:      canvas.Width,
			Height:     canvas.Height,
			OwnerId:    canvas.OwnerID,
			MembersIds: canvas.MembersIDs,
			Privacy:    canvas.Privacy,
			Image:      canvas.Image,
		},
	}, nil
}

func (c *CanvasServer) GetCanvases(ctx context.Context, req *canavasv1.GetCanvasesRequest) (*canavasv1.GetCanvasesResponse, error) {
	if len(req.GetCanvasIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "canvasIDs is required")
	}

	canvases, err := c.canvasService.GetCanvases(ctx, req.CanvasIds)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	var responseCanvases []*canavasv1.Canvas

	for _, canvas := range canvases {
		responseCanvases = append(responseCanvases, &canavasv1.Canvas{
			CanvasId:   canvas.ID,
			Name:       canvas.Name,
			Width:      canvas.Width,
			Height:     canvas.Height,
			OwnerId:    canvas.OwnerID,
			MembersIds: canvas.MembersIDs,
			Privacy:    canvas.Privacy,
			Image:      canvas.Image,
		})
	}

	return &canavasv1.GetCanvasesResponse{
		Canvases: responseCanvases,
	}, nil
}

func (c *CanvasServer) UploadImage(ctx context.Context, req *canavasv1.UploadImageRequest) (*canavasv1.UploadImageResponse, error) {
	if req.GetCanvasId() == "" {
		return nil, status.Error(codes.InvalidArgument, "canvasID is required")
	}

	if len(req.GetImage()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Image is required")
	}

	canvasID, err := c.canvasService.UploadImage(ctx, req.GetCanvasId(), req.GetImage())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &canavasv1.UploadImageResponse{
		CanvasId: canvasID,
	}, nil
}

func (c *CanvasServer) JoinToCanvas(ctx context.Context, req *canavasv1.JoinToCanvasRequest) (*canavasv1.JoinToCanvasResponse, error) {
	if req.GetCanvasId() == "" {
		return nil, status.Error(codes.InvalidArgument, "canvasId is required")
	}
	if req.GetMemberId() == "" {
		return nil, status.Error(codes.InvalidArgument, "memberId is required")
	}

	canvasID, err := c.canvasService.JoinToCanvas(ctx, req.CanvasId, req.MemberId)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &canavasv1.JoinToCanvasResponse{
		CanvasId: canvasID,
	}, nil
}

func (c *CanvasServer) AddToWhiteList(ctx context.Context, req *canavasv1.AddToWhiteListRequest) (*canavasv1.AddToWhiteListResponse, error) {
	if req.GetCanvasId() == "" {
		return nil, status.Error(codes.InvalidArgument, "canvasId is required")
	}
	if req.GetMemberId() == "" {
		return nil, status.Error(codes.InvalidArgument, "memberId is required")
	}

	canvasID, err := c.canvasService.AddToWhiteList(ctx, req.GetCanvasId(), req.GetMemberId())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &canavasv1.AddToWhiteListResponse{
		CanvasId: canvasID,
	}, nil
}

func (c *CanvasServer) UpdateCanvas(ctx context.Context, req *canavasv1.UpdateCanvasRequest) (*canavasv1.UpdateCanvasResponse, error) {
	if req.GetCanvasId() == "" {
		return nil, status.Error(codes.InvalidArgument, "canvasId is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if !slices.Contains([]string{CanvasPrivacyPublic, CanvasPrivacyPrivate, CanvasPrivacyFriends}, req.Privacy) {
		return nil, status.Error(codes.InvalidArgument, "canvas privacy is not included (public, private, friends)")
	}

	canvasID, err := c.canvasService.UpdateCanvas(
		ctx,
		req.GetCanvasId(),
		req.GetName(),
		req.GetPrivacy(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &canavasv1.UpdateCanvasResponse{
		CanvasId: canvasID,
	}, nil
}

func (c *CanvasServer) DeleteCanvas(ctx context.Context, req *canavasv1.DeleteCanvasRequest) (*canavasv1.DeleteCanvasResponse, error) {
	if req.GetCanvasId() == "" {
		return nil, status.Error(codes.InvalidArgument, "canvasId is required")
	}

	canvasID, err := c.canvasService.DeleteCanvas(ctx, req.GetCanvasId())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &canavasv1.DeleteCanvasResponse{
		CanvasId: canvasID,
	}, nil
}

// func (s *Server) GetWhiteList(ctx context.Context, req *canavasv1.GetWhiteListRequest) (*canavasv1.GetWhiteListResponse, error) {
// 	return nil, errors.New("not implemented")
// }
