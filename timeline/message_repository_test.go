package timeline

type MessageRepositoryOnMemory struct {
	data map[string]Message
}

func (r MessageRepositoryOnMemory) FindMessageInTimeline(m Message) (*Message, error) {
	_, found := r.data[m.ToKey()]
	if found {
		return &m, nil
	}
	return nil, nil
}

func (r MessageRepositoryOnMemory) Put(u User, m Message) error {
	r.data[m.ToKey()] = m
	return nil
}

func (r MessageRepositoryOnMemory) Delete(m Message) error {
	delete(r.data, m.ToKey())
	return nil
}
