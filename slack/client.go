package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ara-ta3/slack-timeline/timeline"

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

func (cli SlackClient) ConnectToRTM() (RTMConnection, error) {
	v := url.Values{
		"token": {cli.Token},
	}
	resp, e := http.Get(rtmStartURL + "?" + v.Encode())
	if e != nil {
		return SlackRTMConnection{}, e
	}
	defer resp.Body.Close()
	byteArray, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return SlackRTMConnection{}, e
	}
	res := rtmStartResponse{}
	e = json.Unmarshal(byteArray, &res)
	if e != nil {
		return SlackRTMConnection{}, e
	}
	if !res.OK {
		return SlackRTMConnection{}, fmt.Errorf(res.Error)
	}
	ws, e := websocket.Dial(res.URL, "", origin)
	if e != nil {
		return SlackRTMConnection{}, e
	}
	return SlackRTMConnection{
		ws: ws,
	}, nil
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
