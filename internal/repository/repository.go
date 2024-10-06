package repository

import (
	"context"
	"errors"

	domain "github.com/rkchv/chat/internal/domain/chat"
)

type Repository interface {
	Save(context.Context, *domain.Chat) error
	Get(context.Context, int64) (*domain.Chat, error)
	Update(context.Context, *domain.Chat) error
	Delete(ctx context.Context, id int64) error
}

var (
	// ErrChatNotFound пользователь отсутствует в хранилище
	ErrChatNotFound = errors.New("чат не найден")
)
