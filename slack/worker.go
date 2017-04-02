package slack

import (
	"encoding/json"

	"github.com/pkg/errors"
	"golang.org/x/net/websocket"

	"github.com/ara-ta3/slack-timeline/timeline"
)

type SlackTimelineWorker struct {
	rtmClient RTMClient
}

type RTMClient interface {
	ConnectToRTM() (RTMConnection, error)
}

type RTMConnection interface {
	Read() ([]byte, error)
	Close() error
}

type SlackRTMConnection struct {
	ws *websocket.Conn
}

func (c SlackRTMConnection) Read() ([]byte, error) {
	var msg = make([]byte, 4096)
	n, e := c.ws.Read(msg)
	if e != nil {
		return nil, e
	}
	return msg[:n], nil
}

func (c SlackRTMConnection) Close() error {
	return c.ws.Close()
}

func NewSlackTimelineWorker(rtmClient RTMClient) SlackTimelineWorker {
	return SlackTimelineWorker{
		rtmClient: rtmClient,
	}
}

func (w SlackTimelineWorker) Polling(
	messageChan, deletedMessageChan chan *timeline.Message,
	errorChan chan error,
	endChan chan bool,
) {
	con, e := w.rtmClient.ConnectToRTM()
	if e != nil {
		err := errors.Wrap(e, "failed to connecting to slack rtm")
		errorChan <- err
		return
	}
	defer con.Close()
	prev := make([]byte, 0)
	for {
		received, e := con.Read()
		if e != nil {
			err := errors.Wrap(e, "failed to reading from slack rtm")
			errorChan <- err
		}
		msg := append(prev, received...)
		if !isValidJson(msg) {
			prev = msg
			continue
		}
		prev = make([]byte, 0)
		message := SlackMessage{}
		errOnMessage := json.Unmarshal(msg, &message)
		if errOnMessage != nil {
			continue
		}

		message.Raw = string(msg)
		if message.Type != "message" {
			continue
		}

		if message.IsMessageToPost() {
			m := message.ToInternal()
			messageChan <- &m
		}

		if message.IsDeletedMessage() {
			d := deletedEvent{}
			e := json.Unmarshal([]byte(msg), &d)
			if e != nil {
				continue
			}
			d.Message.ChannelID = d.ChannelID
			m := d.Message.ToInternal()
			deletedMessageChan <- &m
		}
	}
}
