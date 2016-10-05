package timeline

import "log"

type TimelineService struct {
	SlackClient         slackClient
	TimelineChannelID   string
	BlackListChannelIDs []string
	logger              log.Logger
}

func NewTimelineService(slackAPIToken, timelineChannelID string, blackListChannelIDs []string) TimelineService {
	return TimelineService{
		SlackClient:         slackClient{Token: slackAPIToken},
		TimelineChannelID:   timelineChannelID,
		BlackListChannelIDs: blackListChannelIDs,
	}
}

func (service *TimelineService) Run() error {
	messageChan := make(chan *slackMessage)
	errorChan := make(chan error)

	go service.SlackClient.polling(messageChan, errorChan)
	for {
		select {
		case msg := <-messageChan:
			if !service.isTargetMessage(msg) {
				service.logger.Println(msg)
				continue
			}
			e := service.postMessage(msg)
			if e != nil {
				return e
			}
		case e := <-errorChan:
			return e
		default:
			break
		}
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

func (service *TimelineService) postMessage(m *slackMessage) error {
	// TODO cache?
	u, e := service.SlackClient.getUser(m.UserID)
	if e != nil {
		return e
	}

	t := m.Text + " (at <#" + m.ChannelID + ">)"
	// TODO about response
	_, e = service.SlackClient.postMessage(service.TimelineChannelID, t, u.Name, u.Profile.ImageURL)
	return e
}

func isPublic(channelID string) bool {
	return channelID[0:1] == "C"
}
