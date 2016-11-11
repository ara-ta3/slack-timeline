package slack

import (
	"encoding/json"

	"../timeline"
	"github.com/syndtr/goleveldb/leveldb"
)

func NewMessageRepository(timelineChannelID string, s SlackClient, db leveldb.DB) MessageRepositoryOnSlack {
	return MessageRepositoryOnSlack{
		timelineChannelID: timelineChannelID,
		SlackClient:       &s,
		db:                &db,
	}
}

type MessageRepositoryOnSlack struct {
	timelineChannelID string
	SlackClient       *SlackClient
	db                *leveldb.DB
}

func (r MessageRepositoryOnSlack) FindMessageInTimeline(message timeline.Message) (timeline.Message, error) {
	key := message.ToKey()
	data, err := r.db.Get([]byte(key), nil)
	if err != nil {
		return timeline.Message{}, err
	}
	m := SlackMessage{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return timeline.Message{}, err
	}
	return m.ToInternal(), nil
}

func (r MessageRepositoryOnSlack) Put(u timeline.User, m timeline.Message) error {
	if r.alreadExists(m) {
		return nil
	}
	t := m.Text + " (at <#" + m.ChannelID + "> )"
	posted, e := r.SlackClient.postMessage(r.timelineChannelID, t, u.Name, u.ProfileImageURL)
	if e != nil {
		return e
	}
	key := m.ToKey()
	r.db.Put([]byte(key), posted, nil)
	return nil
}

func (r MessageRepositoryOnSlack) Delete(message timeline.Message) error {
	_, e := r.SlackClient.deleteMessage(message.TimeStamp, message.ChannelID)
	return e
}

func (r MessageRepositoryOnSlack) alreadExists(message timeline.Message) bool {
	key := message.ToKey()
	_, err := r.db.Get([]byte(key), nil)
	return err == nil
}
