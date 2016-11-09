package timeline

import (
	"fmt"
	"log"
	"os"
	"testing"

	"../slack"

	"github.com/stretchr/testify/assert"
)

var emptyUserRepository = UserRepositoryOnMemory{data: map[string]slack.User{}}

var emptyMessageRepository = MessageRepositoryOnMemory{data: map[string]slack.SlackMessage{}}

func TestIsTargetReturnFalseWhenReceivedMessageWithTheSameChannelID(t *testing.T) {
	s := NewServiceForTest(emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slack.SlackMessage{
		Type:      "message",
		ChannelID: "CtimelineChannelID",
	}
	assert.Equal(t, false, s.isTargetMessage(&m))
}

func TestIsTargetReturnFalseWhenReceivedMessageFromNotPublicChannel(t *testing.T) {
	s := NewServiceForTest(emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slack.SlackMessage{
		Type:      "message",
		ChannelID: "Phogehoge",
	}
	assert.Equal(t, false, s.isTargetMessage(&m))
}

func TestIsTargetReturnFalseWhenReceivedMessageFromBlacklistedChannel(t *testing.T) {
	s := NewServiceForTest(emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slack.SlackMessage{
		Type:      "message",
		ChannelID: "Caaa",
	}
	assert.Equal(t, false, s.isTargetMessage(&m))
}

func TestIsTargetReturnTrue(t *testing.T) {
	s := NewServiceForTest(emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slack.SlackMessage{
		Type:      "message",
		ChannelID: "Cccc",
	}
	assert.Equal(t, true, s.isTargetMessage(&m))

}

type UserRepositoryOnMemory struct {
	data map[string]slack.User
}

func (r UserRepositoryOnMemory) Get(userID string) (slack.User, error) {
	u, found := r.data[userID]
	if found {
		return u, nil
	}
	return slack.User{}, fmt.Errorf("not found")
}

func (r UserRepositoryOnMemory) Clear() error {
	r.data = map[string]slack.User{}
	return nil
}

type MessageRepositoryOnMemory struct {
	data map[string]slack.SlackMessage
}

func (r MessageRepositoryOnMemory) FindMessageInTimeline(m slack.SlackMessage) (slack.SlackMessage, error) {
	// TODO ホントはm自体ではない
	_, found := r.data[m.ToKey()]
	if found {
		return m, nil
	}
	return slack.SlackMessage{}, fmt.Errorf("not found")
}

func (r MessageRepositoryOnMemory) Put(u slack.User, m slack.SlackMessage) error {
	r.data[m.ToKey()] = m
	return nil
}

func (r MessageRepositoryOnMemory) Delete(m slack.SlackMessage) error {
	delete(r.data, m.ToKey())
	return nil
}

func NewServiceForTest(userRepository UserRepository, messageRepository MessageRepository, t string, bs []string) TimelineService {
	logger := log.New(os.Stdout, "", log.Ldate+log.Ltime+log.Lshortfile)
	return TimelineService{
		SlackClient:         slack.SlackClient{},
		UserRepository:      userRepository,
		MessageRepository:   messageRepository,
		TimelineChannelID:   t,
		BlackListChannelIDs: bs,
		logger:              *logger,
	}
}

func TestTimelineServicePutMessage(t *testing.T) {
	userRepository := UserRepositoryOnMemory{data: map[string]slack.User{
		"userid": slack.User{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]slack.SlackMessage{}}
	s := NewServiceForTest(userRepository, messageRepository, "timelineChannelID", nil)
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
	s := NewServiceForTest(userRepository, messageRepository, "timelineChannelID", nil)
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
