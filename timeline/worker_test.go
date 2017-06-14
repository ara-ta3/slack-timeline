package timeline

type TimelineWorkerMock struct {
	polling func(
		messageChan, deletedMessageChan chan *Message,
		errorChan chan error,
		endChan chan bool,
		userCacheClearChan chan interface{},
	)
}

func (w TimelineWorkerMock) Polling(
	messageChan, deletedMessageChan chan *Message,
	errorChan chan error,
	endChan chan bool,
	userCacheClearChan chan interface{},
) {
	w.polling(messageChan, deletedMessageChan, errorChan, endChan, userCacheClearChan)
}
