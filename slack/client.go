package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/ara-ta3/slack-timeline/timeline"
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

func (m *SlackMessage) IsMessageToPost() bool {
	return m.SubType == "" || m.isFileShare()
}

func (m *SlackMessage) IsDeletedMessage() bool {
	return m.SubType == "message_deleted"
}

func (m *SlackMessage) isFileShare() bool {
	return m.SubType == "file_share"
}

func (m *SlackMessage) ToInternal() timeline.Message {
	return timeline.NewMessage(
		ReplaceIdFormatToName(m.Text),
		m.UserID,
		m.ChannelID,
		m.TimeStamp,
	)
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
	Token            string
	requestWithRetry SlackRetryAble
}

func NewSlackClient(token string, logger *log.Logger) SlackClient {
	return SlackClient{
		Token: token,
		requestWithRetry: SlackRetryAble{
			N:      10,
			logger: logger,
		},
	}
}

func isValidJson(b []byte) bool {
	j := map[string]interface{}{}
	return json.Unmarshal(b, &j) == nil
}

func (cli SlackClient) ConnectToRTM() (RTMConnection, error) {
	v := url.Values{
		"token": {cli.Token},
	}
	r, e := cli.requestWithRetry.GetRequest(rtmStartURL + "?" + v.Encode())

	if e != nil {
		e := errors.Wrap(e, "failed to start rtm connection")
		return SlackRTMConnection{}, e
	}
	defer r.Body.Close()
	byteArray, e := ioutil.ReadAll(r.Body)
	if e != nil {
		e := errors.Wrap(e, fmt.Sprintf("failed read body on starting rtm connection. response: %+v", r))
		return SlackRTMConnection{}, e
	}
	res := rtmStartResponse{}
	e = json.Unmarshal(byteArray, &res)
	if e != nil {
		e := errors.Wrap(e, fmt.Sprintf("failed unmarshal body on starting rtm connection. response: %+v", r))
		return SlackRTMConnection{}, e
	}
	if !res.OK {
		return SlackRTMConnection{}, fmt.Errorf(res.Error)
	}
	ws, e := websocket.Dial(res.URL, "", origin)
	if e != nil {
		e := errors.Wrap(e, fmt.Sprintf("failed dialing to websocket. response: %+v", res))
		return SlackRTMConnection{}, e
	}
	return SlackRTMConnection{
		ws: ws,
	}, nil
}

func (cli *SlackClient) postMessage(channelID, text, userName, iconURL string) ([]byte, error) {
	res, e := cli.requestWithRetry.PostReqest(slackAPIEndpoint+"chat.postMessage", url.Values{
		"token":      {cli.Token},
		"channel":    {channelID},
		"text":       {text},
		"username":   {userName},
		"as_user":    {"false"},
		"icon_url":   {iconURL},
		"link_names": {"0"},
	})
	if e != nil {
		e = errors.Wrap(e, fmt.Sprintf("failed to post message. user: %s, channel: %s. text: %s", userName, channelID, text))
		return nil, e
	}
	defer res.Body.Close()
	byteArray, e := ioutil.ReadAll(res.Body)
	if e != nil {
		e = errors.Wrap(e, fmt.Sprintf("failed read all. response: %+v", res))
		return nil, e
	}
	return byteArray, nil
}

func (cli *SlackClient) getUser(userID string) (*User, error) {
	res, e := cli.requestWithRetry.PostReqest(slackAPIEndpoint+"users.info", url.Values{
		"token": {cli.Token},
		"user":  {userID},
	})
	if e != nil {
		e = errors.Wrap(e, fmt.Sprintf("failed to get user info. user: %s", userID))
		return nil, e
	}
	defer res.Body.Close()
	b, e := ioutil.ReadAll(res.Body)
	if e != nil {
		e = errors.Wrap(e, fmt.Sprintf("failed read all. response: %+v", res))
		return nil, e
	}
	r := userListResponse{}
	e = json.Unmarshal(b, &r)
	if e != nil {
		e = errors.Wrap(e, fmt.Sprintf("failed to Unmarshal response body on get user info. body: %+v", b))
		return nil, e
	}
	if !r.OK {
		return nil, fmt.Errorf(r.Error)
	}
	u := r.User
	return &u, nil
}

func (cli *SlackClient) getAllUsers() ([]User, error) {
	res, e := cli.requestWithRetry.PostReqest(slackAPIEndpoint+"users.list", url.Values{
		"token": {cli.Token},
	})
	if e != nil {
		e = errors.Wrap(e, "failed to get user lists.")
		return nil, e
	}
	defer res.Body.Close()
	b, e := ioutil.ReadAll(res.Body)
	if e != nil {
		e = errors.Wrap(e, fmt.Sprintf("failed read all. response: %+v", res))
		return nil, e
	}
	r := allUserResponse{}
	e = json.Unmarshal(b, &r)
	if e != nil {
		e = errors.Wrap(e, fmt.Sprintf("failed to Unmarshal response body on get users lists. body: %+v, response: %+v", string(b), res))
		return nil, e
	}
	if !r.OK {
		return nil, fmt.Errorf(r.Error)
	}
	return r.Members, nil
}

func (cli *SlackClient) deleteMessage(ts, channel string) ([]byte, error) {
	res, e := cli.requestWithRetry.PostReqest(
		slackAPIEndpoint+"chat.delete",
		url.Values{
			"token":   {cli.Token},
			"ts":      {ts},
			"channel": {channel},
		},
	)
	if e != nil {
		e = errors.Wrap(e, fmt.Sprintf("failed to delete message. ts: %s, channel: %s", ts, channel))
		return nil, e
	}
	defer res.Body.Close()
	byteArray, e := ioutil.ReadAll(res.Body)
	if e != nil {
		e = errors.Wrap(e, fmt.Sprintf("failed read all. response: %+v", res))
		return nil, e
	}
	return byteArray, nil
}
