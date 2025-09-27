package model

import "time"

const (
	StatusWait         = "wait"
	StatusSent         = "sent"
	StatusCanceled     = "canceled"
	StatusNotDelivered = "not delivered"
)

type CreateNotification struct {
	Text     string
	DateTime time.Time
	TgChatId int64
}

type NotificationInRepo struct {
	Uid       string
	Text      string
	DateTime  time.Time
	Status    string
	TgChatId  int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NotificationInResponse struct {
	Uid      string    `json:"id"`
	Text     string    `json:"text"`
	DateTime time.Time `json:"date_time"`
	Status   string    `json:"status"`
}
