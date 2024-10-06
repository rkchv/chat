package services

import (
	"context"
	"errors"

	"github.com/rkchv/auth/pkg/user_v1/auth"
	"github.com/rkchv/chat/lib/logger"
	"github.com/rkchv/chat/lib/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

func (s *Service) Delete(ctx context.Context, chatId int64) error {
	log := logger.GetLogger(ctx)
	var span trace.Span
	ctx, span = tracer.Span(ctx, "Delete")
	defer span.End()

	span.SetAttributes(attribute.Int64("chat_id", chatId))
	span.AddEvent("get user from token")
	tokenUser := auth.UserFromContext(ctx)
	span.SetAttributes(attribute.Int64("user_id", tokenUser.ID))

	span.AddEvent("call auth service")
	canDel, err := s.authService.CanDelete(ctx, tokenUser.ID)
	if err != nil {
		log.Error("failed to call auth service", slog.String("error", err.Error()), slog.String("call_method", "CanDelete"))
		return err
	}

	if !canDel {
		return errors.New("недостаточно прав на удаление")
	}

	span.AddEvent("call repository")
	err = s.chatRepository.Delete(ctx, chatId)
	if err != nil {
		log.Error("failed to delete chat", slog.String("error", err.Error()), slog.Int("chatId", int(chatId)))
	}

	return err
}
