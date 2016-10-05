package timeline

import (
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

	go service.SlackClient.polling(messageChan, errorChan)
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
			_, err := service.db.Get([]byte(key), nil)
			if err == nil || isFirst {
				isFirst = false
				continue
			}
			e := service.postMessage(msg)
			if e != nil {
				service.db.Put([]byte(key), []byte(key), nil)
				return e
			}
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

func (service *TimelineService) postMessage(m *slackMessage) error {
	// TODO cache?
	u, e := service.SlackClient.getUser(m.UserID)
	if e != nil {
		return e
	}

	t := m.Text + " (at <#" + m.ChannelID + ">)"
	// TODO about response
	_, e = service.SlackClient.postMessage(service.TimelineChannelID, t, u.Name, u.Profile.ImageURL)
	return e
}

func isPublic(channelID string) bool {
	return channelID[0:1] == "C"
}
