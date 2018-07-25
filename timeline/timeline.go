package timeline

import (
	"log"

	"fmt"

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
	Get(userID string) (*User, error)
	GetAll() ([]User, error)
	Clear() error
}

type MessageRepository interface {
	FindMessageInTimeline(m Message) (*Message, error)
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
				return e
			}
		case d := <-deletedMessageChan:
			e := s.DeleteFromTimeline(d)
			if e != nil {
				switch e.(type) {
				case MessageNotFoundError:
					// do nothing
				default:
					return e
				}
			}
		case e := <-errorChan:
			return e
		case _ = <-endChan:
			return nil
		case _ = <-userCacheClearChan:
			e := s.UserRepository.Clear()
			if e != nil {
				return e
			}
			s.logger.Printf("User Cache was cleared")
		default:
			break
		}
	}
	return nil
}

func (service *TimelineService) PutToTimeline(m *Message) error {
	if !service.MessageValidator.IsTargetMessage(m) {
		return nil
	}
	u, e := service.UserRepository.Get(m.UserID)
	if e != nil {
		return e
	}
	if u == nil {
		return errors.New(fmt.Sprintf("user not found. id: %s", m.UserID))
	}
	t := service.IDReplacer.Replace(m.Text)
	m.Text = t

	e = service.MessageRepository.Put(*u, *m)
	if e != nil {
		return e
	}
	return nil
}

func (service *TimelineService) DeleteFromTimeline(originMessage *Message) error {
	m, e := service.MessageRepository.FindMessageInTimeline(*originMessage)
	if e != nil {
		return e

	}
	if m == nil {
		return MessageNotFoundError{
			Message: *originMessage,
		}
	}
	e = service.MessageRepository.Delete(*m)
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

type MessageNotFoundError struct {
	Message Message
}

func (e MessageNotFoundError) Error() string {
	return fmt.Sprintf("key of %s is not found\n", e.Message.ToKey())
}
