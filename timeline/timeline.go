package timeline

import (
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

type TimelineService struct {
	slackClient         slackClient
	UserRepository      UserRepository
	MessageRepository   MessageRepository
	TimelineChannelID   string
	BlackListChannelIDs []string
	logger              log.Logger
}

func NewTimelineService(slackAPIToken, timelineChannelID string, blackListChannelIDs []string, db leveldb.DB, logger log.Logger) TimelineService {
	s := slackClient{Token: slackAPIToken}
	return TimelineService{
		slackClient:         s,
		UserRepository:      NewUserRepository(s),
		MessageRepository:   NewMessageRepository(timelineChannelID, s, db),
		TimelineChannelID:   timelineChannelID,
		BlackListChannelIDs: blackListChannelIDs,
		logger:              logger,
	}
}

func (service *TimelineService) Run() error {
	messageChan := make(chan *slackMessage)
	errorChan := make(chan error)
	warnChan := make(chan error)
	deletedMessageChan := make(chan *slackMessage)

	go service.slackClient.polling(messageChan, deletedMessageChan, warnChan, errorChan)
	for {
		select {
		case msg := <-messageChan:
			e := service.PutToTimeline(msg)
			if e != nil {
				service.logger.Printf("%+v\n", e)
			}
		case d := <-deletedMessageChan:
			e := service.DeleteFromTimeline(d)
			if e != nil {
				service.logger.Printf("%+v\n", d)
				service.logger.Printf("%+v\n", e)
			}
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

func (service *TimelineService) PutToTimeline(m *slackMessage) error {
	if !service.isTargetMessage(m) {
		service.logger.Printf("%+v\n", m)
		return nil
	}
	u, e := service.UserRepository.Get(m.UserID)
	if e != nil {
		return e
	}

	e = service.MessageRepository.Put(u, *m)
	if e != nil {
		return e
	}
	return nil
}

func (service *TimelineService) DeleteFromTimeline(originMessage *slackMessage) error {
	m, e := service.MessageRepository.FindMessageInTimeline(*originMessage)
	if e != nil {
		return e
	}
	e = service.MessageRepository.Delete(m)
	if e != nil {
		return e
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

func isPublic(channelID string) bool {
	return channelID[0:1] == "C"
}
