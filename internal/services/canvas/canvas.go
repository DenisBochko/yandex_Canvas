package canvas

import (
	"context"

	"github.com/DenisBochko/yandex_Canvas/internal/domain/models"
)

const (
	CanvasPrivacyPublic  = "public"  // публичный холст
	CanvasPrivacyPrivate = "private" // приватный холст
	CanvasPrivacyFriends = "friends" // холст доступен только друзьям
)

type minioStorage interface {
	SaveCanvas(ctx context.Context, canvasID string, imageData []byte) error
	GetCanvas(ctx context.Context, canvasID string) ([]byte, error)
	DeleteCanvas(ctx context.Context, canvasID string) error
}

type CanvasService struct {
}

func New() *CanvasService {
	return &CanvasService{}
}

func (c *CanvasService) CreateCanvas(ctx context.Context, name string, width int32, height int32, ownerID string, privacy string) (string, error) {
	return "", nil
}

func (c *CanvasService) GetCanvasById(ctx context.Context, id string) (models.Canvas, error) {
	return models.Canvas{}, nil
}

func (c *CanvasService) GetCanvases(ctx context.Context, canvasIDs []string) ([]models.Canvas, error) {
	return []models.Canvas{}, nil
}

func (c *CanvasService) UploadImage(ctx context.Context, canvasID string, image []byte) (string, error) {
	return "", nil
}

func (c *CanvasService) JoinToCanvas(ctx context.Context, canvasID string, userID string) (string, error) {
	return "", nil
}

func (c *CanvasService) AddToWhiteList(ctx context.Context, canvasID string, userID string) (string, error) {
	return "", nil
}

func (c *CanvasService) UpdateCanvas(ctx context.Context, canvasID string, name string, privacy string) (string, error) {
	return "", nil
}

func (c *CanvasService) DeleteCanvas(ctx context.Context, canvasID string) (string, error) {
	return "", nil
}

func (c *CanvasService) GetWhiteList(ctx context.Context, canvasID string) ([]string, error) {
	return []string{}, nil
}
