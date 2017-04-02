package main

import (
	"flag"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/ara-ta3/slack-timeline/slack"
	"github.com/ara-ta3/slack-timeline/timeline"
)

var logger = log.New(os.Stdout, "", log.Ldate+log.Ltime+log.Lshortfile)

func main() {
	filePath := flag.String("c", "config.json", "file path to config.json")
	dbPath := flag.String("db", "db", "path of db for deleting message")
	flag.Parse()
	logger.Printf("filepath: %s\n", *filePath)
	logger.Printf("dbpath: %s\n", *dbPath)
	config, e := ReadConfig(*filePath)
	if e != nil {
		logger.Fatalln(e)
	}

	db, e := leveldb.OpenFile(*dbPath, nil)
	if e != nil {
		logger.Fatalln(e)
	}
	defer db.Close()
	slackClient := slack.NewSlackClient(config.SlackAPIToken, logger)
	worker := slack.NewSlackTimelineWorker(slackClient)
	userRepository := slack.NewUserRepository(slackClient)
	messageRepository := slack.NewMessageRepository(config.TimelineChannelID, slackClient, *db)
	messageValidator := timeline.MessageValidator{
		TimelineChannelID:   config.TimelineChannelID,
		BlackListChannelIDs: config.BlackListChannelIDs,
	}

	service, e := timeline.NewTimelineService(
		worker,
		userRepository,
		messageRepository,
		messageValidator,
		*logger,
	)

	if e != nil {
		log.Fatalf("%+v", e)
	}

	err := service.Run()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
