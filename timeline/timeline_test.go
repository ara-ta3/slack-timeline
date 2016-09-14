package timeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTargetReturnFalseWhenReceivedMessageWithTheSameChannelID(t *testing.T) {
	s := NewTimelineService("token", "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slackMessage{
		Type:      "message",
		ChannelID: "CtimelineChannelID",
	}
	assert.Equal(t, false, s.isTargetMessage(&m))
}

func TestIsTargetReturnFalseWhenReceivedMessageFromNotPublicChannel(t *testing.T) {
	s := NewTimelineService("token", "CtimelineChannelID", []string{"Caaa", "Cbbb"})
	m := slackMessage{
		Type:      "message",
		ChannelID: "Phogehoge",
	}
	assert.Equal(t, false, s.isTargetMessage(&m))
}

func TestIsTargetReturnFalseWhenReceivedMessageFromBlacklistedChannel(t *testing.T) {
	s := NewTimelineService("token", "timelineChannelID", []string{"Caaa", "Cbbb"})
	m := slackMessage{
		Type:      "message",
		ChannelID: "Caaa",
	}
	assert.Equal(t, false, s.isTargetMessage(&m))
}

func TestIsTargetReturnTrue(t *testing.T) {
	s := NewTimelineService("token", "timelineChannelID", []string{"Caaa", "Cbbb"})
	m := slackMessage{
		Type:      "message",
		ChannelID: "Cccc",
	}
	assert.Equal(t, true, s.isTargetMessage(&m))

}
