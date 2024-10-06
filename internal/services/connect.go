package services

import (
	"context"
	"log/slog"

	"github.com/rkchv/chat/lib/logger"

	"github.com/rkchv/chat/internal/services/models"
)

func (s *Service) Connect(ctx context.Context, req models.Connect) error {
	log := logger.GetLogger(ctx)
	ch, err := s.chatRepository.Get(ctx, req.ChatId)
	if err != nil {
		log.Error("failed to get chat", slog.String("error", err.Error()), slog.Any("request", req))
	}

	ch.Connect(req.UserId)
	err = s.chatRepository.Update(ctx, ch)
	if err != nil {
		log.Error("failed to update chat", slog.String("error", err.Error()), slog.Any("request", req))
		return err
	}

	return nil
}
