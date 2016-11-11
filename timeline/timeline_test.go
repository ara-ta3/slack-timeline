package timeline

import (
	"log"
	"os"
	"testing"

	"../slack"

	"github.com/stretchr/testify/assert"
)

var emptyWorker = TimelineWorkerMock{
	polling: func(
		messageChan, deletedMessageChan chan *slack.SlackMessage,
		warnChan, errorChan chan error,
		endChan chan bool,
	) {
	},
}

var emptyUserRepository = UserRepositoryOnMemory{data: map[string]slack.User{}}

var emptyMessageRepository = MessageRepositoryOnMemory{data: map[string]slack.SlackMessage{}}

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
	m := slack.SlackMessage{
		Type:      "message",
		ChannelID: "CtimelineChannelID",
	}
	assert.Equal(t, false, s.MessageValidator.IsTargetMessage(&m))
}

func TestIsTargetReturnFalseWhenReceivedMessageFromNotPublicChannel(t *testing.T) {
	s := NewServiceForTest(emptyWorker, emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slack.SlackMessage{
		Type:      "message",
		ChannelID: "Phogehoge",
	}
	assert.Equal(t, false, s.MessageValidator.IsTargetMessage(&m))
}

func TestIsTargetReturnFalseWhenReceivedMessageFromBlacklistedChannel(t *testing.T) {
	s := NewServiceForTest(emptyWorker, emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slack.SlackMessage{
		Type:      "message",
		ChannelID: "Caaa",
	}
	assert.Equal(t, false, s.MessageValidator.IsTargetMessage(&m))
}

func TestIsTargetReturnTrue(t *testing.T) {
	s := NewServiceForTest(emptyWorker, emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slack.SlackMessage{
		Type:      "message",
		ChannelID: "Cccc",
	}
	assert.Equal(t, true, s.MessageValidator.IsTargetMessage(&m))

}

func TestTimelineServicePutMessage(t *testing.T) {
	userRepository := UserRepositoryOnMemory{data: map[string]slack.User{
		"userid": slack.User{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]slack.SlackMessage{}}
	s := NewServiceForTest(emptyWorker, userRepository, messageRepository, "timelineChannelID", nil)
	m := slack.SlackMessage{
		Raw:       "raw",
		Type:      "message",
		UserID:    "userid",
		Text:      "hogefuga",
		ChannelID: "Cchannel",
		TimeStamp: "ts",
		SubType:   "",
	}
	e := s.PutToTimeline(&m)
	if assert.NoError(t, e) {
		_, found := messageRepository.data[m.ToKey()]
		assert.True(t, found)
	}
}

func TestTimelineServiceDeleteFromTimeline(t *testing.T) {
	userRepository := UserRepositoryOnMemory{data: map[string]slack.User{
		"userid": slack.User{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]slack.SlackMessage{}}
	s := NewServiceForTest(emptyWorker, userRepository, messageRepository, "timelineChannelID", nil)
	m := slack.SlackMessage{
		Raw:       "raw",
		Type:      "message",
		UserID:    "userid",
		Text:      "hogefuga",
		ChannelID: "Cchannel",
		TimeStamp: "ts",
		SubType:   "",
	}
	s.PutToTimeline(&m)
	e := s.DeleteFromTimeline(&m)

	if assert.NoError(t, e) {
		_, found := messageRepository.data[m.ToKey()]
		assert.False(t, found)
	}
}

func TestTimelineServicePutMessageFromWorker(t *testing.T) {
	userRepository := UserRepositoryOnMemory{data: map[string]slack.User{
		"userid": slack.User{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]slack.SlackMessage{}}

	m := slack.SlackMessage{
		Raw:       "raw",
		Type:      "message",
		UserID:    "userid",
		Text:      "hogefuga",
		ChannelID: "Cchannel",
		TimeStamp: "ts",
		SubType:   "",
	}
	polling := func(
		messageChan, deletedMessageChan chan *slack.SlackMessage,
		warnChan, errorChan chan error,
		endChan chan bool,
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
	userRepository := UserRepositoryOnMemory{data: map[string]slack.User{
		"userid": slack.User{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]slack.SlackMessage{}}

	m := slack.SlackMessage{
		Raw:       "raw",
		Type:      "message",
		UserID:    "userid",
		Text:      "hogefuga",
		ChannelID: "Cchannel",
		TimeStamp: "ts",
		SubType:   "",
	}
	polling := func(
		messageChan, deletedMessageChan chan *slack.SlackMessage,
		warnChan, errorChan chan error,
		endChan chan bool,
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
