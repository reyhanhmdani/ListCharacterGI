package entity

import "time"

type Message struct {
	ID          int64     `json:"id"`
	SenderID    int64     `json:"sender_id"`
	ReceiverID  int64     `json:"receiver_id"`
	MessageText string    `json:"message_text"`
	Timestamp   time.Time `json:"timestamp"`
}
