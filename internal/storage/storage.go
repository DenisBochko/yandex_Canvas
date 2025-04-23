package storage

import "errors"

var (
	ErrCanvasExists    = errors.New("canvas already exists")
	ErrInvalidOwnerID  = errors.New("invalid owner UUID")
	ErrInvalidCanvasID = errors.New("invalid canvas UUID")
	ErrAddOwnerToWhiteList = errors.New("cannot add owner to their own whitelist")
)
