package grpc_server

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	userdesc "github.com/rkchv/chat/pkg/chat_v1"
)

// Delete удаляет чат
func (s *Server) Delete(ctx context.Context, req *userdesc.DeleteRequest) (*emptypb.Empty, error) {
	s.CloseChatByID(req.GetId())

	err := s.chatService.Delete(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
