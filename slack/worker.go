package slack

type SlackTimelineWorker struct {
	slackClient SlackClient
}

func NewSlackTimelineWorker(slackClient SlackClient) SlackTimelineWorker {
	return SlackTimelineWorker{
		slackClient: slackClient,
	}
}

func (w SlackTimelineWorker) Polling(
	messageChan, deletedMessageChan chan *SlackMessage,
	warnChan, errorChan chan error,
	endChan chan bool,
) {
	w.slackClient.Polling(messageChan, deletedMessageChan, warnChan, errorChan)
}
