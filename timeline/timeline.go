package timeline

import (
	"log"

	"../slack"
)

type UserRepository interface {
	Get(userID string) (slack.User, error)
	GetAll() ([]slack.User, error)
	Clear() error
}

type MessageRepository interface {
	// TODO slack.SlackMessageじゃなくしたい
	FindMessageInTimeline(m slack.SlackMessage) (slack.SlackMessage, error)
	Put(u slack.User, m slack.SlackMessage) error
	Delete(m slack.SlackMessage) error
}

type TimelineService struct {
	SlackClient       slack.SlackClient
	UserRepository    UserRepository
	MessageRepository MessageRepository
	MessageValidator  MessageValidator
	logger            log.Logger
	IDReplacer        IDReplacer
}

func NewTimelineService(
	slackClient slack.SlackClient,
	userRepository UserRepository,
	messageRepository MessageRepository,
	messageValidator MessageValidator,
	logger log.Logger,
) (TimelineService, error) {
	f := NewIDReplacerFactory(userRepository)
	replacer, e := f.NewReplacer()
	if e != nil {
		return TimelineService{}, e
	}
	return TimelineService{
		SlackClient:       slackClient,
		UserRepository:    userRepository,
		MessageRepository: messageRepository,
		MessageValidator:  messageValidator,
		logger:            logger,
		IDReplacer:        replacer,
	}, nil
}

func (service *TimelineService) Run() error {
	messageChan := make(chan *slack.SlackMessage)
	errorChan := make(chan error)
	warnChan := make(chan error)
	deletedMessageChan := make(chan *slack.SlackMessage)

	go service.SlackClient.Polling(messageChan, deletedMessageChan, warnChan, errorChan)
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

func (service *TimelineService) PutToTimeline(m *slack.SlackMessage) error {
	if !service.MessageValidator.IsTargetMessage(m) {
		service.logger.Printf("%+v\n", m)
		return nil
	}
	u, e := service.UserRepository.Get(m.UserID)
	if e != nil {
		return e
	}
	t := service.IDReplacer.Replace(m.Text)
	m.Text = t

	e = service.MessageRepository.Put(u, *m)
	if e != nil {
		return e
	}
	return nil
}

func (service *TimelineService) DeleteFromTimeline(originMessage *slack.SlackMessage) error {
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

type MessageValidator struct {
	TimelineChannelID   string
	BlackListChannelIDs []string
}

func (v MessageValidator) IsTargetMessage(m *slack.SlackMessage) bool {
	return m.Type == "message" &&
		m.ChannelID != v.TimelineChannelID &&
		isPublic(m.ChannelID) &&
		!contains(v.BlackListChannelIDs, m.ChannelID)
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
