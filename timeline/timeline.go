package timeline

import (
	"encoding/json"
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
	warnChan := make(chan error)
	deletedMessageChan := make(chan *slackMessage)

	go service.SlackClient.polling(messageChan, deletedMessageChan, warnChan, errorChan)
	for {
		select {
		case msg := <-messageChan:
			if !service.isTargetMessage(msg) {
				service.logger.Printf("%+v\n", msg)
				continue
			}
			key := msg.ToKey()
			_, err := service.db.Get([]byte(key), nil)
			if err == nil {
				service.logger.Printf("%+v\n", err)
				continue
			}
			m, e := service.postMessage(msg)
			if e != nil {
				service.logger.Printf("%+v\n", msg)
				service.logger.Printf("%+v\n", e)
			}
			service.db.Put([]byte(key), m, nil)
		case d := <-deletedMessageChan:
			key := d.ToKey()
			data, err := service.db.Get([]byte(key), nil)
			if err != nil {
				service.logger.Printf("%+v\n", d)
				service.logger.Printf("%+v\n", err)
			}
			m := slackMessage{}
			err = json.Unmarshal(data, &m)
			if err != nil {
				service.logger.Printf("%+v\n", d)
				service.logger.Printf("%+v\n", data)
				service.logger.Printf("%+v\n", err)
			}
			service.deleteMessage(&m)
		case e := <-warnChan:
			service.logger.Printf("%+v\n", e)
		case e := <-errorChan:
			return e

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

	t := m.Text + " (at <#" + m.ChannelID + "> )"
	return service.SlackClient.postMessage(service.TimelineChannelID, t, u.Name, u.Profile.ImageURL)
}

func (service *TimelineService) deleteMessage(m *slackMessage) ([]byte, error) {
	return service.SlackClient.deleteMessage(m.TimeStamp, m.ChannelID)
}

func isPublic(channelID string) bool {
	return channelID[0:1] == "C"
}
