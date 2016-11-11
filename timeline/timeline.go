package timeline

import (
	"log"

	"../slack"
)

type TimelineWorker interface {
	Polling(
		messageChan, deletedMessageChan chan *slack.SlackMessage,
		warnChan, errorChan chan error,
		endChan chan bool,
	)
}

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
	TimelineWorker    TimelineWorker
	UserRepository    UserRepository
	MessageRepository MessageRepository
	MessageValidator  MessageValidator
	logger            log.Logger
	IDReplacer        IDReplacer
}

func NewTimelineService(
	timelineWorker TimelineWorker,
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
		TimelineWorker:    timelineWorker,
		UserRepository:    userRepository,
		MessageRepository: messageRepository,
		MessageValidator:  messageValidator,
		logger:            logger,
		IDReplacer:        replacer,
	}, nil
}

func (s *TimelineService) Run() error {
	messageChan := make(chan *slack.SlackMessage)
	deletedMessageChan := make(chan *slack.SlackMessage)
	errorChan := make(chan error)
	warnChan := make(chan error)
	endChan := make(chan bool)

	go s.TimelineWorker.Polling(
		messageChan,
		deletedMessageChan,
		warnChan,
		errorChan,
		endChan,
	)
	for {
		select {
		case msg := <-messageChan:
			e := s.PutToTimeline(msg)
			if e != nil {
				s.logger.Printf("%+v\n", e)
			}
		case d := <-deletedMessageChan:
			e := s.DeleteFromTimeline(d)
			if e != nil {
				s.logger.Printf("%+v\n", d)
				s.logger.Printf("%+v\n", e)
			}
		case e := <-warnChan:
			s.logger.Printf("%+v\n", e)
		case e := <-errorChan:
			return e
		case _ = <-endChan:
			return nil
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
