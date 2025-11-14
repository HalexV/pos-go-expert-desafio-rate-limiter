package limit_entity

import (
	"context"
	"time"
)

type Limit struct {
	Id      string
	FreeAt  *time.Time
	LastAt  time.Time
	Counter int32
}

type LimitEntityRepository interface {
	CreateLimit(ctx context.Context, limit *Limit) error
	GetLimitById(ctx context.Context, id string) (*Limit, error)
	UpdateLimitById(ctx context.Context, id string, limit *Limit) error
}
