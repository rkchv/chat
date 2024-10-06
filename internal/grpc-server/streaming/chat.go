package streaming

import (
	"sync"

	chatdesc "github.com/rkchv/chat/pkg/chat_v1"
)

type Chat interface {
	ID() int64
	Connect(userID int64, stream chatdesc.ChatV1_ConnectServer)
	Disconnect(userID int64)
	IsEmpty() bool
	Close()
	AddMessage(msg *chatdesc.Message)
	BroadcastMessages()
}

// chat Чат
type chat struct {
	id          int64
	messages    chan *chatdesc.Message
	connections map[int64]chatdesc.ChatV1_ConnectServer
	m           sync.RWMutex
}

// NewChat Создает новый чат
func NewChat(id int64) Chat {
	return &chat{
		id:          id,
		connections: make(map[int64]chatdesc.ChatV1_ConnectServer),
		messages:    make(chan *chatdesc.Message),
	}
}

func (c *chat) ID() int64 {
	return c.id
}

// Connect Подключает пользователя (его стрим) к чату
func (c *chat) Connect(userID int64, stream chatdesc.ChatV1_ConnectServer) {
	c.m.Lock()
	defer c.m.Unlock()
	c.connections[userID] = stream
}

// Disconnect Отключает пользователя (его стрим) от чата
func (c *chat) Disconnect(userID int64) {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.connections, userID)
}

// IsEmpty Проверяет есть ли в чате еще активные соединения (стримы)
func (c *chat) IsEmpty() bool {
	c.m.RLock()
	defer c.m.RUnlock()
	return len(c.connections) == 0
}

// Close Закрывает чат и освобождает ресурсы
func (c *chat) Close() {
	close(c.messages)
	c.m.Lock()
	defer c.m.Unlock()
	c.connections = nil
}

func (c *chat) AddMessage(msg *chatdesc.Message) {
	c.messages <- msg
}

func (c *chat) BroadcastMessages() {
	for {
		select {
		case msg, ok := <-c.messages:
			if !ok {
				return
			}
			c.m.RLock()
			for _, stream := range c.connections {
				_ = stream.Send(msg)
			}
			c.m.RUnlock()
		default:
			return
		}
	}
}
