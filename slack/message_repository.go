package slack

import (
	"encoding/json"

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

func (r MessageRepositoryOnSlack) FindMessageInTimeline(message SlackMessage) (SlackMessage, error) {
	key := message.ToKey()
	data, err := r.db.Get([]byte(key), nil)
	if err != nil {
		return SlackMessage{}, err
	}
	m := SlackMessage{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return SlackMessage{}, err
	}
	return m, nil
}

func (r MessageRepositoryOnSlack) Put(u User, m SlackMessage) error {
	if r.alreadExists(m) {
		return nil
	}
	t := m.Text + " (at <#" + m.ChannelID + "> )"
	posted, e := r.SlackClient.postMessage(r.timelineChannelID, t, u.Name, u.Profile.ImageURL)
	if e != nil {
		return e
	}
	key := m.ToKey()
	r.db.Put([]byte(key), posted, nil)
	return nil
}

func (r MessageRepositoryOnSlack) Delete(message SlackMessage) error {
	_, e := r.SlackClient.deleteMessage(message.TimeStamp, message.ChannelID)
	return e
}

func (r MessageRepositoryOnSlack) alreadExists(message SlackMessage) bool {
	key := message.ToKey()
	_, err := r.db.Get([]byte(key), nil)
	return err == nil
}
