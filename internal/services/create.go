package services

import (
	"context"

	"github.com/rkchv/chat/lib/logger"
	"golang.org/x/exp/slog"

	"github.com/rkchv/chat/internal/domain/chat"
)

func (s *Service) Create(ctx context.Context) (int64, error) {
	log := logger.GetLogger(ctx)
	ch := chat.NewChat()
	err := s.chatRepository.Save(ctx, &ch)

	if err != nil {
		log.Error("failed to create chat", slog.String("error", err.Error()))
	}

	return ch.Id, err
}
