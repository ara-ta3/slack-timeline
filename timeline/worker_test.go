package timeline

type TimelineWorkerMock struct {
	polling func(
		messageChan, deletedMessageChan chan *Message,
		warnChan, errorChan chan error,
		endChan, restartChan chan bool,
	)
}

func (w TimelineWorkerMock) Polling(
	messageChan, deletedMessageChan chan *Message,
	warnChan, errorChan chan error,
	endChan, restartChan chan bool,
) {
	w.polling(messageChan, deletedMessageChan, warnChan, errorChan, endChan, restartChan)
}
