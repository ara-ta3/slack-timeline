# SlackTimeline
===

[![Build Status](https://travis-ci.org/ara-ta3/slacktimeline.svg?branch=master)](https://travis-ci.org/ara-ta3/slacktimeline)

SlackTimeline sends all messages in public channels to a specific channel. e.g. #timeline.  

## Config  

Please see config.sample.json.  

```
{
    "slackApiToken": "",
    "timelineChannelID": "",
    "blackListChannelIDs": []
}
```

* slackApiToken
  * Slack API Token or Hubot API Token
* timelineChannelID
  * The ID of the channel to post all public channel's messages. 
  * Something like `C01234567`
* blackListChannelIDs
  * The some ID of the channels from which you don't want to post to the "TimelineChannel".
  * Something like `[C00000000, C00000001]`
    * If the settings like this, messages from the channel of "C00000000" and "C00000001" never post to the "TimelineChannel".
