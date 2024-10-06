package services

import (
	"context"

	"github.com/rkchv/chat/internal/domain/chat"
	"github.com/rkchv/chat/internal/repository"
	"github.com/rkchv/chat/internal/services/models"
)

type ChatService interface {
	Create(ctx context.Context) (int64, error)
	Connect(ctx context.Context, req models.Connect) error
	Delete(ctx context.Context, chatId int64) error
	Get(ctx context.Context, chatId int64) (chat.Chat, error)
}

type AuthServiceClient interface {
	CanDelete(ctx context.Context, userID int64) (bool, error)
}

type Service struct {
	chatRepository repository.Repository
	authService    AuthServiceClient
}

func NewService(chatRepository repository.Repository, authClient AuthServiceClient) *Service {
	return &Service{
		chatRepository: chatRepository,
		authService:    authClient,
	}
}
