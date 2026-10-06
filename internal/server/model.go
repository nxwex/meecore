package server

import (
	"errors"
	"time"
)

var ErrInstanceNotFound = errors.New("instance not found")

type Instance struct {
	ID          int64
	NodeID      int64
	Name        string
	Template    string
	ContainerID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsDeleted   bool
}
