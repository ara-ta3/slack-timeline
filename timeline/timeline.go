package timeline

import (
	"log"

	"github.com/pkg/errors"
)

type TimelineWorker interface {
	Polling(
		messageChan, deletedMessageChan chan *Message,
		errorChan chan error,
		endChan chan bool,
		userCacheClearChan chan interface{},
	)
}

type UserRepository interface {
	Get(userID string) (User, error)
	GetAll() ([]User, error)
	Clear() error
}

type MessageRepository interface {
	FindMessageInTimeline(m Message) (Message, error)
	Put(u User, m Message) error
	Delete(m Message) error
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
	messageChan := make(chan *Message)
	deletedMessageChan := make(chan *Message)
	errorChan := make(chan error)
	endChan := make(chan bool)
	userCacheClearChan := make(chan interface{})

	go s.TimelineWorker.Polling(
		messageChan,
		deletedMessageChan,
		errorChan,
		endChan,
		userCacheClearChan,
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
		case e := <-errorChan:
			return e
		case _ = <-endChan:
			return nil
		case _ = <-userCacheClearChan:
			s.logger.Printf("User Cache will be cleared")
			s.UserRepository.Clear()
		default:
			break
		}
	}
	return nil
}

func (service *TimelineService) PutToTimeline(m *Message) error {
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

func (service *TimelineService) DeleteFromTimeline(originMessage *Message) error {
	m, e := service.MessageRepository.FindMessageInTimeline(*originMessage)
	if e != nil {
		e = errors.Wrap(e, "failed to find message in timeline")
		return e
	}
	e = service.MessageRepository.Delete(m)
	if e != nil {
		e = errors.Wrap(e, "failed to delete message in timeline")
		return e
	}
	return nil
}

type MessageValidator struct {
	TimelineChannelID   string
	BlackListChannelIDs []string
}

func (v MessageValidator) IsTargetMessage(m *Message) bool {
	return m.ChannelID != v.TimelineChannelID &&
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
