package models

type AddToWhiteListMessage struct {
	CanvasID   string
	CanvasName string // для красоты
	OwnerID    string
	UserId     string // id пользователя, которого добавляют
}
