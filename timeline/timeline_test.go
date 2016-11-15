package timeline

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var emptyWorker = TimelineWorkerMock{
	polling: func(
		messageChan, deletedMessageChan chan *Message,
		warnChan, errorChan chan error,
		endChan, restartChan chan bool,
	) {
		endChan <- true
	},
}

var emptyUserRepository = UserRepositoryOnMemory{data: map[string]User{}}

var emptyMessageRepository = MessageRepositoryOnMemory{data: map[string]Message{}}

func NewServiceForTest(
	worker TimelineWorker,
	userRepository UserRepository,
	messageRepository MessageRepository,
	t string,
	bs []string,
) TimelineService {
	logger := log.New(os.Stdout, "", log.Ldate+log.Ltime+log.Lshortfile)
	v := MessageValidator{
		TimelineChannelID:   t,
		BlackListChannelIDs: bs,
	}
	r, _ := NewTimelineService(worker, userRepository, messageRepository, v, *logger)
	return r
}

func TestIsTargetReturnFalseWhenReceivedMessageWithTheSameChannelID(t *testing.T) {
	s := NewServiceForTest(emptyWorker, emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := Message{
		ChannelID: "CtimelineChannelID",
	}
	assert.Equal(t, false, s.MessageValidator.IsTargetMessage(&m))
}

func TestIsTargetReturnFalseWhenReceivedMessageFromNotPublicChannel(t *testing.T) {
	s := NewServiceForTest(emptyWorker, emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := Message{
		ChannelID: "Phogehoge",
	}
	assert.Equal(t, false, s.MessageValidator.IsTargetMessage(&m))
}

func TestIsTargetReturnFalseWhenReceivedMessageFromBlacklistedChannel(t *testing.T) {
	s := NewServiceForTest(emptyWorker, emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := Message{
		ChannelID: "Caaa",
	}
	assert.Equal(t, false, s.MessageValidator.IsTargetMessage(&m))
}

func TestIsTargetReturnTrue(t *testing.T) {
	s := NewServiceForTest(emptyWorker, emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := Message{
		ChannelID: "Cccc",
	}
	assert.Equal(t, true, s.MessageValidator.IsTargetMessage(&m))

}

func TestTimelineServicePutMessage(t *testing.T) {
	userRepository := UserRepositoryOnMemory{data: map[string]User{
		"userid": User{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]Message{}}
	s := NewServiceForTest(emptyWorker, userRepository, messageRepository, "timelineChannelID", nil)
	m := Message{
		Text:      "hogefuga",
		UserID:    "userid",
		ChannelID: "Cchannel",
		TimeStamp: "ts",
	}
	e := s.PutToTimeline(&m)
	if assert.NoError(t, e) {
		_, found := messageRepository.data[m.ToKey()]
		assert.True(t, found)
	}
}

func TestTimelineServiceDeleteFromTimeline(t *testing.T) {
	userRepository := UserRepositoryOnMemory{data: map[string]User{
		"userid": User{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]Message{}}
	s := NewServiceForTest(emptyWorker, userRepository, messageRepository, "timelineChannelID", nil)
	m := Message{
		Text:      "hogefuga",
		UserID:    "userid",
		ChannelID: "Cchannel",
		TimeStamp: "ts",
	}
	s.PutToTimeline(&m)
	e := s.DeleteFromTimeline(&m)

	if assert.NoError(t, e) {
		_, found := messageRepository.data[m.ToKey()]
		assert.False(t, found)
	}
}

func TestTimelineServicePutMessageFromWorker(t *testing.T) {
	userRepository := UserRepositoryOnMemory{data: map[string]User{
		"userid": User{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]Message{}}

	m := Message{
		Text:      "hogefuga",
		UserID:    "userid",
		ChannelID: "Cchannel",
		TimeStamp: "ts",
	}
	polling := func(
		messageChan, deletedMessageChan chan *Message,
		warnChan, errorChan chan error,
		endChan, restartChan chan bool,
	) {
		messageChan <- &m
		endChan <- true
	}
	worker := TimelineWorkerMock{polling: polling}
	s := NewServiceForTest(worker, userRepository, messageRepository, "timelineChannelID", nil)
	s.Run()
	actual, found := messageRepository.data[m.ToKey()]
	assert.True(t, found)
	assert.Equal(t, m.Text, actual.Text)
}

func TestTimelineServiceDeleteFromTimelineFromWorker(t *testing.T) {
	userRepository := UserRepositoryOnMemory{data: map[string]User{
		"userid": User{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]Message{}}

	m := Message{
		Text:      "hogefuga",
		ChannelID: "Cchannel",
		TimeStamp: "ts",
	}
	polling := func(
		messageChan, deletedMessageChan chan *Message,
		warnChan, errorChan chan error,
		endChan, restartChan chan bool,
	) {
		messageChan <- &m
		deletedMessageChan <- &m
		endChan <- true
	}
	worker := TimelineWorkerMock{polling: polling}
	s := NewServiceForTest(worker, userRepository, messageRepository, "timelineChannelID", nil)
	s.Run()
	_, found := messageRepository.data[m.ToKey()]
	assert.False(t, found)
}
