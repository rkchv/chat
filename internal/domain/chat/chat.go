package chat

import (
	"slices"
	"time"
)

type Chat struct {
	Id        int64
	UserIds   []int64
	CreatedAt time.Time
}

func NewChat() Chat {
	return Chat{
		CreatedAt: time.Now(),
	}
}

func (c *Chat) Connect(userId int64) {
	if !slices.Contains(c.UserIds, userId) {
		c.UserIds = append(c.UserIds, userId)
	}
}
