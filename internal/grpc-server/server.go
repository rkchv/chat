package grpc_server

import (
	"sync"
	"time"

	"github.com/rkchv/chat/internal/grpc-server/metrics"
	"github.com/rkchv/chat/internal/grpc-server/streaming"
	"github.com/rkchv/chat/internal/services"
	chatdesc "github.com/rkchv/chat/pkg/chat_v1"
)

type Server struct {
	chatdesc.UnimplementedChatV1Server
	chatService    *services.Service
	connectedChats map[int64]streaming.Chat
	m              sync.RWMutex
	metrics        *metrics.Metrics
	garbageCycle   time.Duration
	chatExpiration time.Duration
}

func NewServer(srv *services.Service, garbageCycle time.Duration, chatExpired time.Duration) *Server {
	return &Server{
		chatService:    srv,
		metrics:        metrics.NewMetrics(),
		connectedChats: make(map[int64]streaming.Chat),
		garbageCycle:   garbageCycle,
		chatExpiration: chatExpired,
	}
}

func (s *Server) NewChat(chatId int64) streaming.Chat {
	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.connectedChats[chatId]; ok {
		return streaming.NewChat(chatId)
	}

	newChat := streaming.NewChat(chatId)
	s.connectedChats[chatId] = newChat
	s.metrics.IncreaseChats()

	return newChat
}

func (s *Server) CloseChat(ch streaming.Chat) {
	ch.Close()
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.connectedChats, ch.ID())
	s.metrics.DecreaseChats()
}

func (s *Server) CloseChatByID(chatId int64) {
	s.m.RLock()
	ch, ok := s.connectedChats[chatId]
	s.m.RUnlock()
	if ok {
		s.CloseChat(ch)
	}
}
