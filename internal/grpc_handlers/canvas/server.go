package canvas

import (
	"context"
	"errors"
	"slices"

	"github.com/DenisBochko/yandex_Canvas/internal/domain/models"
	"github.com/DenisBochko/yandex_Canvas/internal/storage"
	canavasv1 "gitlab.crja72.ru/golang/2025/spring/course/projects/go6/contracts/gen/go/canvas"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TOOD: GetWhiteList
// TODO: Сделать человеческую обработку ошибок

type CanvasService interface {
	CreateCanvas(ctx context.Context, name string, width int32, height int32, ownerID string, privacy string) (string, error) // возвращает id созданного канваса

	GetCanvasById(ctx context.Context, canvasID string) (*models.Canvas, error)
	GetCanvasByIdNoImage(ctx context.Context, canvasID string) (*models.Canvas, error)
	GetCanvases(ctx context.Context, canvasIDs []string) ([]models.Canvas, error)
	GetCanvasesNoImage(ctx context.Context, canvasIDs []string) ([]models.Canvas, error)
	GetCanvasesByUserId(ctx context.Context, userID string) ([]models.Canvas, error)

	UploadImage(ctx context.Context, canvasID string, image []byte) (string, error) // возвращает id загруженного изображения

	// Одобрение по почте (отправляем owner`у на почту)
	JoinToCanvas(ctx context.Context, canvasID string, userID string) (string, error)   // возвращает id канваса
	AddToWhiteList(ctx context.Context, canvasID string, userID string) (string, error) // возвращает id канваса

	UpdateCanvas(ctx context.Context, canvasID string, name string, privacy string) (string, error) // возвращает id обновлённого канваса
	DeleteCanvas(ctx context.Context, canvasID string) (string, error)                              // возвращает id удалённого канваса
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
		if errors.Is(err, storage.ErrInvalidOwnerID) {
			return nil, status.Error(codes.InvalidArgument, "invalid owner UUID")
		}
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

	canvas, err := c.canvasService.GetCanvasByIdNoImage(ctx, req.GetCanvasId())
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
		},
	}, nil
}

func (c *CanvasServer) GetImageById(ctx context.Context, req *canavasv1.GetImageByIdRequest) (*canavasv1.GetImageByIdResponse, error) {
	if req.GetCanvasId() == "" {
		return nil, status.Error(codes.InvalidArgument, "canvasID is required")
	}

	canvas, err := c.canvasService.GetCanvasById(ctx, req.GetCanvasId())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &canavasv1.GetImageByIdResponse{
		Image: canvas.Image,
	}, nil
}

func (c *CanvasServer) GetImageByIds(ctx context.Context, req *canavasv1.GetImageByIdsRequest) (*canavasv1.GetImageByIdsResponse, error) {
	if len(req.GetCanvasIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "canvasIDs is required")
	}

	canvases, err := c.canvasService.GetCanvases(ctx, req.CanvasIds)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	var imagesResponse [][]byte

	for _, image := range canvases {
		imagesResponse = append(imagesResponse, image.Image)
	}

	return &canavasv1.GetImageByIdsResponse{
		Images: imagesResponse,
	}, nil
}

func (c *CanvasServer) GetCanvasesByUserId(ctx context.Context, req *canavasv1.GetCanvasesByUserIdRequest) (*canavasv1.GetCanvasesByUserIdResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "userID is required")
	}

	canvases, err := c.canvasService.GetCanvasesByUserId(ctx, req.GetUserId())
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
		})
	}

	return &canavasv1.GetCanvasesByUserIdResponse{
		Canvases: responseCanvases,
	}, nil
}

func (c *CanvasServer) GetCanvases(ctx context.Context, req *canavasv1.GetCanvasesRequest) (*canavasv1.GetCanvasesResponse, error) {
	if len(req.GetCanvasIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "canvasIDs is required")
	}

	canvases, err := c.canvasService.GetCanvasesNoImage(ctx, req.CanvasIds)
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
		return nil, err
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
	// user id должно совпадать с id owner`а canvas, к которому хотим присоедениться
	userID, verified, ok := extractUserInfo(ctx)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user ID or verification info missing")
	}
	if verified == "false" {
		return nil, status.Error(codes.Unauthenticated, "user not verified")
	}

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

func extractUserInfo(ctx context.Context) (uid string, verified string, ok bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", false
	}

	uids := md.Get("uid")
	verifieds := md.Get("verified")

	if len(uids) == 0 || len(verifieds) == 0 {
		return "", "", false
	}

	return uids[0], verifieds[0], true
}
