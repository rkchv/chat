package grpc_server

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	userdesc "github.com/rkchv/chat/pkg/chat_v1"
)

// Create создает чат
func (s *Server) Create(ctx context.Context, _ *emptypb.Empty) (*userdesc.CreateResponse, error) {
	id, err := s.chatService.Create(ctx)
	if err != nil {
		return nil, err
	}

	newChat := s.NewChat(id)

	go func() {
		lastNonEmpty := time.Now()
		t := time.NewTicker(s.garbageCycle)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				if newChat.IsEmpty() {
					if time.Since(lastNonEmpty) > s.chatExpiration {
						s.CloseChat(newChat)
						return
					}
				} else {
					lastNonEmpty = time.Now()
				}
			default:
				newChat.BroadcastMessages()
			}
		}
	}()

	return &userdesc.CreateResponse{Id: id}, nil
}
