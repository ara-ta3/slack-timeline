package timeline

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

type TimelineService struct {
	SlackClient         slackClient
	TimelineChannelID   string
	BlackListChannelIDs []string
	db                  leveldb.DB
	logger              log.Logger
}

func NewTimelineService(slackAPIToken, timelineChannelID string, blackListChannelIDs []string, db leveldb.DB, logger log.Logger) TimelineService {
	return TimelineService{
		SlackClient:         slackClient{Token: slackAPIToken},
		TimelineChannelID:   timelineChannelID,
		BlackListChannelIDs: blackListChannelIDs,
		db:                  db,
		logger:              logger,
	}
}

func (service *TimelineService) Run() error {
	messageChan := make(chan *slackMessage)
	errorChan := make(chan error)
	deletedMessageChan := make(chan *slackMessage)

	go service.SlackClient.polling(messageChan, errorChan, deletedMessageChan)
	isFirst := false
	for {
		select {
		case msg := <-messageChan:
			if msg.Type == "hello" {
				isFirst = true
			}
			if !service.isTargetMessage(msg) {
				service.logger.Println(msg)
				continue
			}
			key := msg.ChannelID + "-" + msg.TimeStamp
			fmt.Println(key)
			_, err := service.db.Get([]byte(key), nil)
			if err == nil || isFirst {
				isFirst = false
				continue
			}
			m, e := service.postMessage(msg)
			if e != nil {
				service.logger.Println(msg)
				service.logger.Println(e)
			}
			service.db.Put([]byte(key), m, nil)
		case e := <-errorChan:
			return e
		case d := <-deletedMessageChan:
			key := d.ChannelID + "-" + d.TimeStamp
			data, err := service.db.Get([]byte(key), nil)
			if err != nil {
				service.logger.Println(d)
				service.logger.Println(err)
			}
			m := slackMessage{}
			err = json.Unmarshal(data, &m)
			if err != nil {
				service.logger.Println(d)
				service.logger.Println(err)
			}
			service.deleteMessage(&m)
		default:
			break
		}
	}
	return nil
}

func (service *TimelineService) isTargetMessage(m *slackMessage) bool {
	return m.Type == "message" &&
		m.ChannelID != service.TimelineChannelID &&
		isPublic(m.ChannelID) &&
		!contains(service.BlackListChannelIDs, m.ChannelID)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (service *TimelineService) postMessage(m *slackMessage) ([]byte, error) {
	// TODO cache?
	u, e := service.SlackClient.getUser(m.UserID)
	if e != nil {
		return nil, e
	}

	t := m.Text + " (at <#" + m.ChannelID + ">)"
	// TODO about response
	return service.SlackClient.postMessage(service.TimelineChannelID, t, u.Name, u.Profile.ImageURL)
}

func (service *TimelineService) deleteMessage(m *slackMessage) ([]byte, error) {
	return service.SlackClient.deleteMessage(m.TimeStamp, m.ChannelID)
}

func isPublic(channelID string) bool {
	return channelID[0:1] == "C"
}
