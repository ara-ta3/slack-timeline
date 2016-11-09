package timeline

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var emptyUserRepository = UserRepositoryOnMemory{data: map[string]user{}}

var emptyMessageRepository = MessageRepositoryOnMemory{data: map[string]slackMessage{}}

func TestIsTargetReturnFalseWhenReceivedMessageWithTheSameChannelID(t *testing.T) {
	s := NewServiceForTest(emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slackMessage{
		Type:      "message",
		ChannelID: "CtimelineChannelID",
	}
	assert.Equal(t, false, s.isTargetMessage(&m))
}

func TestIsTargetReturnFalseWhenReceivedMessageFromNotPublicChannel(t *testing.T) {
	s := NewServiceForTest(emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slackMessage{
		Type:      "message",
		ChannelID: "Phogehoge",
	}
	assert.Equal(t, false, s.isTargetMessage(&m))
}

func TestIsTargetReturnFalseWhenReceivedMessageFromBlacklistedChannel(t *testing.T) {
	s := NewServiceForTest(emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slackMessage{
		Type:      "message",
		ChannelID: "Caaa",
	}
	assert.Equal(t, false, s.isTargetMessage(&m))
}

func TestIsTargetReturnTrue(t *testing.T) {
	s := NewServiceForTest(emptyUserRepository, emptyMessageRepository, "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slackMessage{
		Type:      "message",
		ChannelID: "Cccc",
	}
	assert.Equal(t, true, s.isTargetMessage(&m))

}

type UserRepositoryOnMemory struct {
	data map[string]user
}

func (r UserRepositoryOnMemory) Get(userID string) (user, error) {
	u, found := r.data[userID]
	if found {
		return u, nil
	}
	return user{}, fmt.Errorf("not found")
}

func (r UserRepositoryOnMemory) Clear() error {
	r.data = map[string]user{}
	return nil
}

type MessageRepositoryOnMemory struct {
	data map[string]slackMessage
}

func (r MessageRepositoryOnMemory) FindMessageInTimeline(m slackMessage) (slackMessage, error) {
	// TODO ホントはm自体ではない
	_, found := r.data[m.ToKey()]
	if found {
		return m, nil
	}
	return slackMessage{}, fmt.Errorf("not found")
}

func (r MessageRepositoryOnMemory) Put(u user, m slackMessage) error {
	r.data[m.ToKey()] = m
	return nil
}

func (r MessageRepositoryOnMemory) Delete(m slackMessage) error {
	delete(r.data, m.ToKey())
	return nil
}

func NewServiceForTest(userRepository UserRepository, messageRepository MessageRepository, t string, bs []string) TimelineService {
	logger := log.New(os.Stdout, "", log.Ldate+log.Ltime+log.Lshortfile)
	return TimelineService{
		slackClient:         slackClient{},
		UserRepository:      userRepository,
		MessageRepository:   messageRepository,
		TimelineChannelID:   t,
		BlackListChannelIDs: bs,
		logger:              *logger,
	}
}

func TestTimelineServicePutMessage(t *testing.T) {
	userRepository := UserRepositoryOnMemory{data: map[string]user{
		"userid": user{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]slackMessage{}}
	s := NewServiceForTest(userRepository, messageRepository, "timelineChannelID", nil)
	m := slackMessage{
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
	userRepository := UserRepositoryOnMemory{data: map[string]user{
		"userid": user{},
	}}
	messageRepository := MessageRepositoryOnMemory{data: map[string]slackMessage{}}
	s := NewServiceForTest(userRepository, messageRepository, "timelineChannelID", nil)
	m := slackMessage{
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
