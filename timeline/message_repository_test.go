package timeline

import "fmt"

type MessageRepositoryOnMemory struct {
	data map[string]Message
}

func (r MessageRepositoryOnMemory) FindMessageInTimeline(m Message) (Message, error) {
	// TODO ホントはm自体ではない
	_, found := r.data[m.ToKey()]
	if found {
		return m, nil
	}
	return Message{}, fmt.Errorf("not found")
}

func (r MessageRepositoryOnMemory) Put(u User, m Message) error {
	r.data[m.ToKey()] = m
	return nil
}

func (r MessageRepositoryOnMemory) Delete(m Message) error {
	delete(r.data, m.ToKey())
	return nil
}
