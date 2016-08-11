package main

import (
	"encoding/json"
	"log"
	"os"

	"./slack"
)

var slackAPIToken string

var timelineChannelID string

var slackClient slack.SlackClient

var messageChan = make(chan []byte)

var errorChan = make(chan error)

func main() {
	timelineChannelID = os.Getenv("SLACK_TIMELINE_CHANNEL_ID")
	slackAPIToken = os.Getenv("SLACK_TOKEN")
	slackClient = slack.SlackClient{Token: slackAPIToken}

	go slackClient.Polling(messageChan, errorChan)
	for {
		select {
		case msg := <-messageChan:
			e := postMessage(msg)
			if e != nil {
				log.Fatal(e)
			}
		case e := <-errorChan:
			log.Fatal(e)
		default:
			break
		}
	}
}

func postMessage(msg []byte) error {
	m, e := toMessage(msg)
	if e != nil {
		return e
	}
	switch m.Type {
	case "message":
		if m.ChannelID == timelineChannelID {
			return nil
		}
		if !isPublic(m.ChannelID) {
			return nil
		}
		t := m.Text + " (at <#" + m.ChannelID + ">)"
		// TODO about response
		_, e := slackClient.PostMessage(timelineChannelID, t, m.UserID)
		if e != nil {
			return e
		}
	}
	return nil
}

func isPublic(channelID string) bool {
	return channelID[0:1] == "C"
}

func toMessage(msg []byte) (*slack.SlackMessage, error) {
	message := slack.SlackMessage{}
	err := json.Unmarshal(msg, &message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}
