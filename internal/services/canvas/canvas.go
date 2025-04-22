package canvas

import (
	"context"
	"fmt"

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
	GetCanvasesByUserId(ctx context.Context, userID string) ([]*models.InternalCanvas, error)
	SetImageUrl(ctx context.Context, canvasID string, imageURL string) (string, error)
	AddToWhiteList(ctx context.Context, canvasID string, userID string) (string, error)
	GetByIDs(ctx context.Context, canvasIDs []string) ([]*models.InternalCanvas, error)
	Update(ctx context.Context, canvasID string, name string, privacy string) (string, error)
	Delete(ctx context.Context, canvasID string) (string, error)
}

type kafkaTransport interface {
	SendAddToWhiteListMessage(ctx context.Context, message models.AddToWhiteListMessage) error
}

type CanvasService struct {
	minioStorage    minioStorage
	postgresStorage postgresStorage
	kafkaTransport  kafkaTransport
}

func New(postgresStorage postgresStorage, minioStorage minioStorage, kafkaTransport kafkaTransport) *CanvasService {
	return &CanvasService{
		minioStorage:    minioStorage,
		postgresStorage: postgresStorage,
		kafkaTransport:  kafkaTransport,
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

func (c *CanvasService) GetCanvasByIdNoImage(ctx context.Context, canvasID string) (*models.Canvas, error) {
	internalCanvas, err := c.postgresStorage.GetByID(ctx, canvasID)
	if err != nil {
		return nil, nil
	}

	return &models.Canvas{
		ID:         internalCanvas.ID,
		Name:       internalCanvas.Name,
		Width:      internalCanvas.Width,
		Height:     internalCanvas.Height,
		MembersIDs: internalCanvas.MembersIDs,
		Privacy:    internalCanvas.Privacy,
		Image:      []byte{}, // Затычка, т.к. функция не ходит в minio для получение самого канваса
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

func (c *CanvasService) GetCanvasesNoImage(ctx context.Context, canvasIDs []string) ([]models.Canvas, error) {
	internalCanvases, err := c.postgresStorage.GetByIDs(ctx, canvasIDs)
	if err != nil {
		return nil, err
	}

	var canvases []models.Canvas

	for _, canvas := range internalCanvases {
		canvases = append(canvases, models.Canvas{
			ID:         canvas.ID,
			Name:       canvas.Name,
			Width:      canvas.Width,
			Height:     canvas.Height,
			OwnerID:    canvas.OwnerID,
			MembersIDs: canvas.MembersIDs,
			Privacy:    canvas.Privacy,
			Image:      []byte{}, // Затычка, т.к. функция не ходит в minio для получение самого канваса
		})
	}

	return canvases, nil
}

func (c *CanvasService) GetCanvasesByUserId(ctx context.Context, userID string) ([]models.Canvas, error) {
	internalCanvases, err := c.postgresStorage.GetCanvasesByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	var canvases []models.Canvas

	for _, canvas := range internalCanvases {
		canvases = append(canvases, models.Canvas{
			ID:         canvas.ID,
			Name:       canvas.Name,
			Width:      canvas.Width,
			Height:     canvas.Height,
			OwnerID:    canvas.OwnerID,
			MembersIDs: canvas.MembersIDs,
			Privacy:    canvas.Privacy,
			Image:      []byte{}, // Затычка, т.к. функция не ходит в minio для получение самого канваса
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

func (c *CanvasService) JoinToCanvas(ctx context.Context, canvasID string, userID string) (string, error) { // userID - человек, который хочет присодиниться
	internalCanvas, err := c.postgresStorage.GetByID(ctx, canvasID)
	if err != nil {
		return "", err
	}

	// Собираем сообщение для отправки в kafka
	message := models.AddToWhiteListMessage{
		CanvasID:   canvasID,
		CanvasName: internalCanvas.Name,
		OwnerID:    internalCanvas.OwnerID,
		UserId:     userID,
	}

	// отправляем
	if err = c.kafkaTransport.SendAddToWhiteListMessage(ctx, message); err != nil {
		return "", fmt.Errorf("failed to send verification message: %w", err)
	}

	return canvasID, nil
}

func (c *CanvasService) AddToWhiteList(ctx context.Context, canvasID string, userID string) (string, error) {
	_, err := c.postgresStorage.AddToWhiteList(ctx, canvasID, userID)
	if err != nil {
		return "", err
	}

	return canvasID, nil
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
