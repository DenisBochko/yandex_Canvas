package postgresstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/DenisBochko/yandex_Canvas/internal/domain/models"
	"github.com/DenisBochko/yandex_Canvas/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

// Конструктор Storage
func New(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Save(ctx context.Context, canvas models.Canvas) (string, error) {
	canvasID, err := uuid.Parse(canvas.ID)
	if err != nil {
		return "", fmt.Errorf("invalid canvas UUID: %w", err)
	}

	ownerID, err := uuid.Parse(canvas.OwnerID)
	if err != nil {
		return "", storage.ErrInvalidOwnerID
	}

	membersIDs := make([]uuid.UUID, 0, len(canvas.MembersIDs))
	for _, idStr := range canvas.MembersIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return "", fmt.Errorf("invalid member UUID: %w", err)
		}
		membersIDs = append(membersIDs, id)
	}

	// сохраняем без картинки, она лежит в MinIO
	_, err = s.db.Exec(ctx, `
		INSERT INTO canvases(canvas_id, name, width, height, owner_id, members_ids, privacy, image)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
	`, canvasID, canvas.Name, canvas.Width, canvas.Height, ownerID, membersIDs, canvas.Privacy, "") // image = ""

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case "23505":
				return "", storage.ErrCanvasExists
			default:
				return "", fmt.Errorf("failed to save canvas: %w", err)
			}
		}
		return "", fmt.Errorf("failed to execute insert: %w", err)
	}

	return canvas.ID, nil
}

func (s *Storage) GetCanvasesByUserId(ctx context.Context, userID string) ([]*models.InternalCanvas, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT canvas_id, name, width, height, owner_id, members_ids, privacy, image, created_at
		FROM canvases
		WHERE owner_id = $1 OR $1 = ANY(members_ids)
	`, uid)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var canvases []*models.InternalCanvas

	for rows.Next() {
		var (
			canvasID   uuid.UUID
			name       string
			width      int32
			height     int32
			ownerID    uuid.UUID
			membersIDs []uuid.UUID
			privacy    string
			image      string
			created_at time.Time
		)

		err := rows.Scan(&canvasID, &name, &width, &height, &ownerID, &membersIDs, &privacy, &image, &created_at)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		memberStrs := make([]string, len(membersIDs))
		for i, m := range membersIDs {
			memberStrs[i] = m.String()
		}

		canvases = append(canvases, &models.InternalCanvas{
			ID:         canvasID.String(),
			Name:       name,
			Width:      width,
			Height:     height,
			OwnerID:    ownerID.String(),
			MembersIDs: memberStrs,
			Privacy:    privacy,
			ImageURL:   image,
			CreatedAt:  created_at,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return canvases, nil
}

func (s *Storage) GetByID(ctx context.Context, canvasID string) (*models.InternalCanvas, error) {
	id, err := uuid.Parse(canvasID)
	if err != nil {
		return nil, storage.ErrInvalidCanvasID
	}

	var (
		name       string
		width      int32
		height     int32
		ownerID    uuid.UUID
		membersIDs []uuid.UUID
		privacy    string
		image      string
		createdAt  time.Time
	)

	err = s.db.QueryRow(ctx, `
		SELECT name, width, height, owner_id, members_ids, privacy, image, created_at
		FROM canvases
		WHERE canvas_id = $1
	`, id).Scan(&name, &width, &height, &ownerID, &membersIDs, &privacy, &image, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get canvas: %w", err)
	}

	memberStrs := make([]string, len(membersIDs))
	for i, m := range membersIDs {
		memberStrs[i] = m.String()
	}

	return &models.InternalCanvas{
		ID:         canvasID,
		Name:       name,
		Width:      width,
		Height:     height,
		OwnerID:    ownerID.String(),
		MembersIDs: memberStrs,
		Privacy:    privacy,
		ImageURL:   image,
		CreatedAt:  createdAt,
	}, nil
}

func (s *Storage) SetImageUrl(ctx context.Context, canvasID string, imageURL string) (string, error) {
	id, err := uuid.Parse(canvasID)
	if err != nil {
		return "", fmt.Errorf("invalid canvas UUID: %w", err)
	}

	_, err = s.db.Exec(ctx, `
		UPDATE canvases
		SET image = $1
		WHERE canvas_id = $2
	`, imageURL, id)
	if err != nil {
		return "", fmt.Errorf("failed to update image URL: %w", err)
	}

	return imageURL, nil
}

func (s *Storage) GetByIDs(ctx context.Context, canvasIDs []string) ([]*models.InternalCanvas, error) {
	var canvases []*models.InternalCanvas

	ids := make([]uuid.UUID, 0, len(canvasIDs))
	for _, id := range canvasIDs {
		iID, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid canvas UUID: %w", err)
		}
		ids = append(ids, iID)
	}

	rows, err := s.db.Query(ctx, `
		SELECT canvas_id, name, width, height, owner_id, members_ids, privacy, image, created_at
		FROM canvases
		WHERE canvas_id = ANY($1)
	`, ids)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case "22P02":
				return nil, fmt.Errorf("invalid input syntax for UUID: %w", err)
			default:
				return nil, fmt.Errorf("pg error while getting canvases: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to get canvases: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			canvasID   uuid.UUID
			name       string
			width      int32
			height     int32
			ownerID    uuid.UUID
			membersIDs []uuid.UUID
			privacy    string
			imageURL   string
			created_at time.Time
		)

		if err := rows.Scan(&canvasID, &name, &width, &height, &ownerID, &membersIDs, &privacy, &imageURL, &created_at); err != nil {
			return nil, fmt.Errorf("failed to scan canvas: %w", err)
		}

		memberStrs := make([]string, len(membersIDs))
		for i, m := range membersIDs {
			memberStrs[i] = m.String()
		}

		canvases = append(canvases, &models.InternalCanvas{
			ID:         canvasID.String(),
			Name:       name,
			Width:      width,
			Height:     height,
			OwnerID:    ownerID.String(),
			MembersIDs: memberStrs,
			Privacy:    privacy,
			ImageURL:   imageURL,
			CreatedAt:  created_at,
		})
	}

	return canvases, nil
}

func (s *Storage) Update(ctx context.Context, canvasID string, name string, privacy string) (string, error) {
	id, err := uuid.Parse(canvasID)
	if err != nil {
		return "", fmt.Errorf("invalid canvas UUID: %w", err)
	}

	_, err = s.db.Exec(ctx, `
		UPDATE canvases
		SET name = $1, privacy = $2
		WHERE canvas_id = $3
	`, name, privacy, id)
	if err != nil {
		return "", fmt.Errorf("failed to update canvas: %w", err)
	}

	return canvasID, nil
}

func (s *Storage) Delete(ctx context.Context, canvasID string) (string, error) {
	id, err := uuid.Parse(canvasID)
	if err != nil {
		return "", fmt.Errorf("invalid canvas UUID: %w", err)
	}

	_, err = s.db.Exec(ctx, "DELETE FROM canvases WHERE canvas_id = $1", id)
	if err != nil {
		return "", fmt.Errorf("failed to delete canvas: %w", err)
	}

	return canvasID, nil
}

func (s *Storage) AddToWhiteList(ctx context.Context, canvasID string, userID string) (string, error) {
	cid, err := uuid.Parse(canvasID)
	if err != nil {
		return "", fmt.Errorf("invalid canvas ID: %w", err)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// Получаем owner_id для проверки, чтобы owner не добавился в members своего же канваса
	var ownerID uuid.UUID
	err = s.db.QueryRow(ctx, `
		SELECT owner_id FROM canvases WHERE canvas_id = $1
	`, cid).Scan(&ownerID)
	if err != nil {
		return "", fmt.Errorf("failed to get owner ID: %w", err)
	}

	// Проверяем, не совпадает ли userID с ownerID
	if uid == ownerID {
		return "", storage.ErrAddOwnerToWhiteList
	}

	sqlResponse, err := s.db.Exec(ctx, `
		UPDATE canvases
		SET members_ids = array_append(members_ids, $2)
		WHERE canvas_id = $1 AND NOT ($2 = ANY(members_ids))
	`, cid, uid)

	if err != nil {
		return "", fmt.Errorf("failed to update canvas: %w", err)
	}

	if sqlResponse.RowsAffected() == 0 {
		return "", fmt.Errorf("user already in whitelist or canvas not found")
	}

	return canvasID, nil
}
