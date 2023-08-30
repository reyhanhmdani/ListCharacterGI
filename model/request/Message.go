package request

import (
	"ListCharacterGI/model/entity"
	"time"
)

type SendMessageRequest struct {
	ReceiverID  int64  `json:"receiver_id"`
	MessageText string `json:"message_text"`
}

type ResponseSendMessage struct {
	Message        string  `json:"message"`
	CreatedMessage Message `json:"created_message"`
}

type Message struct {
	ID          int64     `json:"id"`
	SenderID    int64     `json:"sender_id"`
	ReceiverID  int64     `json:"receiver_id"`
	MessageText string    `json:"message_text"`
	Timestamp   time.Time `json:"timestamp"`
}

type ResponseGetMessages struct {
	Message  string    `json:"message"`
	Messages []Message `json:"messages"`
}

func ConvertToResponseMessages(messages []entity.Message) []Message {
	var responseMessages []Message
	for _, msg := range messages {
		responseMsg := Message{
			ID:          msg.ID,
			SenderID:    msg.SenderID,
			ReceiverID:  msg.ReceiverID,
			MessageText: msg.MessageText,
			Timestamp:   msg.Timestamp,
		}
		responseMessages = append(responseMessages, responseMsg)
	}
	return responseMessages
}
