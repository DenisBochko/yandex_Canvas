package models

type Canvas struct {
	ID         string
	Name       string
	Width      int32
	Height     int32
	OwnerID    string   // ID полльзователя, которому принадлежит холст
	MembersIDs []string // ID пользователей, которые имеют доступ к холсту
	Privacy    string   // приватность канваса (public, private, friends)
	Image      []byte   // изображение канваса в формате PNG
}
