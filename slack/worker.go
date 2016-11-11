package slack

import "../timeline"

type SlackTimelineWorker struct {
	slackClient SlackClient
}

func NewSlackTimelineWorker(slackClient SlackClient) SlackTimelineWorker {
	return SlackTimelineWorker{
		slackClient: slackClient,
	}
}

func (w SlackTimelineWorker) Polling(
	messageChan, deletedMessageChan chan *timeline.Message,
	warnChan, errorChan chan error,
	endChan chan bool,
) {
	w.slackClient.Polling(messageChan, deletedMessageChan, warnChan, errorChan)
}
