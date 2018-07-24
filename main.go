package main

import (
	"flag"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/ara-ta3/slack-timeline/logger"
	"github.com/ara-ta3/slack-timeline/slack"
	"github.com/ara-ta3/slack-timeline/timeline"
)

var stdoutLogger = log.New(os.Stdout, "", log.Ldate+log.Ltime+log.Lshortfile)

func main() {
	filePath := flag.String("c", "config.json", "file path to config.json")
	dbPath := flag.String("db", "db", "path of db for deleting message")
	flag.Parse()
	stdoutLogger.Printf("filepath: %s\n", *filePath)
	stdoutLogger.Printf("dbpath: %s\n", *dbPath)
	config, e := ReadConfig(*filePath)
	if e != nil {
		stdoutLogger.Fatalf("%+v\n", e)
	}

	reporter, e := logger.NewReporter(config.Sentry.DSN)
	if e != nil {
		stdoutLogger.Fatalf("%+v\n", e)
	}
	db, e := leveldb.OpenFile(*dbPath, nil)
	if e != nil {
		stdoutLogger.Fatalln(e)
	}
	defer db.Close()
	slackClient := slack.NewSlackClient(config.SlackAPIToken, stdoutLogger)
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
		*stdoutLogger,
	)

	if e != nil {
		reporter.Report(e)
		log.Fatalf("%+v\n", e)
	}

	err := service.Run()
	if err != nil {
		reporter.Report(err)
		log.Fatalf("%+v\n", err)
	}
}
