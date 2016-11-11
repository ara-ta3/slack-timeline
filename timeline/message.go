package timeline

import "fmt"

type Message struct {
	Text      string
	UserID    string
	ChannelID string
	TimeStamp string
}

func (m Message) ToKey() string {
	return fmt.Sprintf("%s-%s", m.ChannelID, m.TimeStamp)
}

func NewMessage(text, userID, channelID, timestamp string) Message {
	return Message{
		Text:      text,
		UserID:    userID,
		ChannelID: channelID,
		TimeStamp: timestamp,
	}
}
