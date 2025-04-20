package canvas

import (
	"context"

	"github.com/DenisBochko/yandex_Canvas/internal/domain/models"
	"github.com/google/uuid"
)

const (
	CanvasPrivacyPublic  = "public"  // публичный холст
	CanvasPrivacyPrivate = "private" // приватный холст
	CanvasPrivacyFriends = "friends" // холст доступен только друзьям
)

type minioStorage interface {
	SaveCanvas(ctx context.Context, canvasID string, imageData []byte) (string, error)
	GetCanvas(ctx context.Context, canvasID string) ([]byte, error)
	DeleteCanvas(ctx context.Context, canvasID string) error
}

type postgresStorage interface {
	Save(ctx context.Context, canvas models.Canvas) (string, error)
	GetByID(ctx context.Context, canvasID string) (*models.InternalCanvas, error)
	SetImageUrl(ctx context.Context, canvasID string, imageURL string) (string, error)
	GetByIDs(ctx context.Context, canvasIDs []string) ([]*models.InternalCanvas, error)
	Update(ctx context.Context, canvasID string, name string, privacy string) (string, error)
	Delete(ctx context.Context, canvasID string) (string, error)
}

type CanvasService struct {
	minioStorage    minioStorage
	postgresStorage postgresStorage
}

func New(postgresStorage postgresStorage, minioStorage minioStorage) *CanvasService {
	return &CanvasService{
		minioStorage:    minioStorage,
		postgresStorage: postgresStorage,
	}
}

func (c *CanvasService) CreateCanvas(ctx context.Context, name string, width int32, height int32, ownerID string, privacy string) (string, error) {
	uuid := uuid.New().String()

	canvas := models.Canvas{
		ID:         uuid,
		Name:       name,
		Width:      width,
		Height:     height,
		OwnerID:    ownerID,
		MembersIDs: []string{},
		Privacy:    privacy,
		Image:      []byte{}, // Image не будем сохранять в бд на этапе создания
	}

	canvasID, err := c.postgresStorage.Save(ctx, canvas)
	if err != nil {
		return "", err
	}

	return canvasID, nil
}

func (c *CanvasService) GetCanvasById(ctx context.Context, canvasID string) (*models.Canvas, error) {
	internalCanvas, err := c.postgresStorage.GetByID(ctx, canvasID)
	if err != nil {
		return nil, nil
	}

	canvasImage, err := c.minioStorage.GetCanvas(ctx, canvasID)
	if err != nil {
		canvasImage = []byte{}
	}

	return &models.Canvas{
		ID:         internalCanvas.ID,
		Name:       internalCanvas.Name,
		Width:      internalCanvas.Width,
		Height:     internalCanvas.Height,
		MembersIDs: internalCanvas.MembersIDs,
		Privacy:    internalCanvas.Privacy,
		Image:      canvasImage,
	}, nil
}

func (c *CanvasService) GetCanvases(ctx context.Context, canvasIDs []string) ([]models.Canvas, error) {
	internalCanvases, err := c.postgresStorage.GetByIDs(ctx, canvasIDs)
	if err != nil {
		return nil, err
	}

	var canvases []models.Canvas

	for _, canvas := range internalCanvases {
		canvasImage, err := c.minioStorage.GetCanvas(ctx, canvas.ID)
		if err != nil {
			canvasImage = []byte{}
		}

		canvases = append(canvases, models.Canvas{
			ID:         canvas.ID,
			Name:       canvas.Name,
			Width:      canvas.Width,
			Height:     canvas.Height,
			OwnerID:    canvas.OwnerID,
			MembersIDs: canvas.MembersIDs,
			Privacy:    canvas.Privacy,
			Image:      canvasImage,
		})
	}

	return canvases, nil
}

func (c *CanvasService) UploadImage(ctx context.Context, canvasID string, image []byte) (string, error) {
	// Сохраняем в minio
	imageUrl, err := c.minioStorage.SaveCanvas(ctx, canvasID, image)
	if err != nil {
		return "", err
	}

	// Обновляем поле image в postgres
	_, err = c.postgresStorage.SetImageUrl(ctx, canvasID, imageUrl)
	if err != nil {
		return "", err
	}

	return canvasID, nil
}

func (c *CanvasService) JoinToCanvas(ctx context.Context, canvasID string, userID string) (string, error) {
	return "", nil
}

func (c *CanvasService) AddToWhiteList(ctx context.Context, canvasID string, userID string) (string, error) {
	return "", nil
}

func (c *CanvasService) UpdateCanvas(ctx context.Context, canvasID string, name string, privacy string) (string, error) {
	id, err := c.postgresStorage.Update(ctx, canvasID, name, privacy)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *CanvasService) DeleteCanvas(ctx context.Context, canvasID string) (string, error) {
	// удаляем canvas из бд
	id, err := c.postgresStorage.Delete(ctx, canvasID)
	if err != nil {
		return "", err
	}

	// удаляем canvas из minio
	err = c.minioStorage.DeleteCanvas(ctx, canvasID)
	if err != nil {
		return "", err
	}

	return id, nil
}

// func (c *CanvasService) GetWhiteList(ctx context.Context, canvasID string) ([]string, error) {
// 	return []string{}, nil
// }
