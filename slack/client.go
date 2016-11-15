package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"../timeline"
	"github.com/pkg/errors"

	"golang.org/x/net/websocket"
)

var rtmStartURL = "https://slack.com/api/rtm.start"

var slackAPIEndpoint = "https://slack.com/api/"

var origin = "http://localhost"

type rtmStartResponse struct {
	OK    bool   `json:"ok"`
	URL   string `json:"url"`
	Error string `json:"error"`
}

type deletedEvent struct {
	ChannelID string       `json:"channel"`
	Message   SlackMessage `json:"previous_message"`
}

type SlackMessage struct {
	Raw       string `json:"-"`
	Type      string `json:"type"`
	UserID    string `json:"user"`
	Text      string `json:"text"`
	ChannelID string `json:"channel"`
	TimeStamp string `json:"ts"`
	SubType   string `json:"subtype"`
}

func (m *SlackMessage) ToInternal() timeline.Message {
	return timeline.NewMessage(m.Text, m.UserID, m.ChannelID, m.TimeStamp)
}

type userListResponse struct {
	OK    bool   `json:"ok"`
	User  User   `json:"user"`
	Error string `json:"error"`
}

type allUserResponse struct {
	OK      bool   `json:"ok"`
	Members []User `json:"members"`
	Error   string `json:"error"`
}

type User struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Profile profile `json:"profile"`
}

func (u User) ToInternal() timeline.User {
	return timeline.NewUser(u.ID, u.Name, u.Profile.ImageURL)
}

type profile struct {
	ImageURL string `json:"image_48"`
}

type channelCreated struct {
	Type    string  `json:"type"`
	Channel channel `json:"channel"`
}

type channel struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Creator string `json:"creator"`
}

type SlackClient struct {
	Token string
}

func isValidJson(b []byte) bool {
	j := map[string]interface{}{}
	return json.Unmarshal(b, &j) == nil
}

func (cli *SlackClient) connectToRTM() (*websocket.Conn, error) {
	v := url.Values{
		"token": {cli.Token},
	}
	resp, e := http.Get(rtmStartURL + "?" + v.Encode())
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	byteArray, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, e
	}
	res := rtmStartResponse{}
	e = json.Unmarshal(byteArray, &res)
	if e != nil {
		return nil, e
	}
	if !res.OK {
		return nil, fmt.Errorf(res.Error)
	}
	ws, e := websocket.Dial(res.URL, "", origin)
	if e != nil {
		return nil, e
	}
	return ws, nil
}

func receive(ws *websocket.Conn) ([]byte, error) {
	var msg = make([]byte, 4096)
	n, e := ws.Read(msg)
	if e != nil {
		return nil, e
	}
	return msg[:n], nil
}

func (cli *SlackClient) Polling(
	messageChan, deletedMessageChan chan *timeline.Message,
	warnChan, errorChan chan error,
	restartChan chan bool,
) {
	ws, e := cli.connectToRTM()
	if e != nil {
		errorChan <- e
		return
	}
	defer ws.Close()
	prev := make([]byte, 0)
	for {
		received, e := receive(ws)
		if e != nil {
			errorChan <- e
		}
		msg := append(prev, received...)
		if !isValidJson(msg) {
			prev = msg
			continue
		}
		prev = make([]byte, 0)
		message := SlackMessage{}
		errOnMessage := json.Unmarshal(msg, &message)

		if errOnMessage != nil {
			event := channelCreated{}
			errOnEvent := json.Unmarshal(msg, &event)
			if errOnEvent != nil && errOnMessage != nil {
				warnChan <- errors.Wrap(errOnMessage, fmt.Sprintf("failed to unmarshal to message. json: '%s'", msg))
				warnChan <- errors.Wrap(errOnEvent, fmt.Sprintf("failed to unmarshal to channel created. json: '%s'", msg))
				continue
			}
			restartChan <- true
			return
		}

		message.Raw = string(msg)
		if message.Type != "message" {
			warnChan <- fmt.Errorf("not message: '%+v'", message)
			continue
		}

		switch message.SubType {
		case "":
			m := message.ToInternal()
			messageChan <- &m
		case "message_deleted":
			d := deletedEvent{}
			e := json.Unmarshal([]byte(msg), &d)
			if e != nil {
				warnChan <- errors.Wrap(e, fmt.Sprintf("failed to unmarshal to deleted event. json: '%s'", msg))
				continue
			}
			d.Message.ChannelID = d.ChannelID
			m := d.Message.ToInternal()
			deletedMessageChan <- &m
		}
	}
}

func (cli *SlackClient) postMessage(channelID, text, userName, iconURL string) ([]byte, error) {
	res, e := http.PostForm(slackAPIEndpoint+"chat.postMessage", url.Values{
		"token":      {cli.Token},
		"channel":    {channelID},
		"text":       {text},
		"username":   {userName},
		"as_user":    {"false"},
		"icon_url":   {iconURL},
		"link_names": {"0"},
	})
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	byteArray, e := ioutil.ReadAll(res.Body)
	if e != nil {
		return nil, e
	}
	return byteArray, nil
}

func (cli *SlackClient) getUser(userID string) (*User, error) {
	res, e := http.PostForm(slackAPIEndpoint+"users.info", url.Values{
		"token": {cli.Token},
		"user":  {userID},
	})
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	b, e := ioutil.ReadAll(res.Body)
	if e != nil {
		return nil, e
	}
	r := userListResponse{}
	e = json.Unmarshal(b, &r)
	if e != nil {
		return nil, e
	}
	if !r.OK {
		return nil, fmt.Errorf(r.Error)
	}
	u := r.User
	return &u, nil
}

func (cli *SlackClient) getAllUsers() ([]User, error) {
	res, e := http.PostForm(slackAPIEndpoint+"users.list", url.Values{
		"token": {cli.Token},
	})
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	b, e := ioutil.ReadAll(res.Body)
	if e != nil {
		return nil, e
	}
	r := allUserResponse{}
	e = json.Unmarshal(b, &r)
	if e != nil {
		return nil, e
	}
	if !r.OK {
		return nil, fmt.Errorf(r.Error)
	}
	return r.Members, nil
}

func (cli *SlackClient) deleteMessage(ts, channel string) ([]byte, error) {
	res, e := http.PostForm(slackAPIEndpoint+"chat.delete", url.Values{
		"token":   {cli.Token},
		"ts":      {ts},
		"channel": {channel},
	})
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	byteArray, e := ioutil.ReadAll(res.Body)
	if e != nil {
		return nil, e
	}
	return byteArray, nil
}
