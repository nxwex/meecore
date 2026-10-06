package node

import (
	"context"
	"errors"
	"net/netip"
	"time"
)

var ErrNotFound = errors.New("node not found")

type Node struct {
	ID        int64
	Name      string
	Hostname  string
	IPAddress netip.Addr
	CreatedAt time.Time
}

type NodeRepository interface {
	Get(ctx context.Context, id int64) (*Node, error)
	GetAll(ctx context.Context) ([]Node, error)
}
