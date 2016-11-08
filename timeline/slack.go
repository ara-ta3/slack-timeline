package timeline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

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
	Message   slackMessage `json:"previous_message"`
}

type slackMessage struct {
	Type      string `json:"type"`
	UserID    string `json:"user"`
	Text      string `json:"text"`
	ChannelID string `json:"channel"`
	TimeStamp string `json:"ts"`
	SubType   string `json:"subtype"`
}

type userListResponse struct {
	OK    bool   `json:"ok"`
	User  user   `json:"user"`
	Error string `json:"error"`
}

type user struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Profile profile `json:"profile"`
}

type profile struct {
	ImageURL string `json:"image_48"`
}

type slackClient struct {
	Token string
}

func (cli *slackClient) connectToRTM() (*websocket.Conn, error) {
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

func (cli *slackClient) polling(messageChan chan *slackMessage, errorChan chan error, deletedMessageChan chan *slackMessage) {
	ws, e := cli.connectToRTM()
	if e != nil {
		errorChan <- e
		return
	}
	defer ws.Close()
	for {
		var msg = make([]byte, 4096)
		n, e := ws.Read(msg)
		if e != nil {
			errorChan <- e
		} else {
			fmt.Println(string(msg[:n]))
			message := slackMessage{}
			err := json.Unmarshal(msg[:n], &message)
			if err != nil {
				fmt.Printf("%+v\n", errors.Wrap(err, fmt.Sprintf("failed to unmarshal. json: '%s'", string(msg[:n]))))
			} else if message.SubType == "" {
				messageChan <- &message
			} else if message.SubType == "message_deleted" {
				d := deletedEvent{}
				e := json.Unmarshal(msg[:n], &d)
				if e != nil {
					fmt.Printf("%+v\n", errors.Wrap(err, fmt.Sprintf("failed to unmarshal to deleted event. json: '%s'", string(msg[:n]))))
				} else {
					d.Message.ChannelID = d.ChannelID
					deletedMessageChan <- &d.Message
				}
			}
		}
	}
}

func (cli *slackClient) postMessage(channelID, text, userName, iconURL string) ([]byte, error) {
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

func (cli *slackClient) getUser(userID string) (*user, error) {
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

func (cli *slackClient) deleteMessage(ts, channel string) ([]byte, error) {
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
