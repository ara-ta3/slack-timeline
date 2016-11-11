package timeline

import (
	"fmt"

	"../slack"
)

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
