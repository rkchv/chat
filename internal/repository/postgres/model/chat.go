package model

import (
	"time"
)

type ChatDTO struct {
	Id        int32
	UserIds   []int64
	CreatedAt time.Time
}
