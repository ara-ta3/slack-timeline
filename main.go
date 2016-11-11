package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"

	"./slack"
	"./timeline"
)

func main() {
	filePath := flag.String("c", "config.json", "file path to config.json")
	dbPath := flag.String("db", "db", "path of db for deleting message")
	flag.Parse()
	fmt.Printf("filepath: %s\n", *filePath)
	fmt.Printf("dbpath: %s\n", *dbPath)
	config, e := ReadConfig(*filePath)
	if e != nil {
		log.Fatalln(e)
	}

	db, e := leveldb.OpenFile(*dbPath, nil)
	if e != nil {
		log.Fatalln(e)
	}
	defer db.Close()

	slackClient := slack.SlackClient{Token: config.SlackAPIToken}
	worker := slack.NewSlackTimelineWorker(slackClient)
	userRepository := slack.NewUserRepository(slackClient)
	messageRepository := slack.NewMessageRepository(config.TimelineChannelID, slackClient, *db)
	messageValidator := timeline.MessageValidator{
		TimelineChannelID:   config.TimelineChannelID,
		BlackListChannelIDs: config.BlackListChannelIDs,
	}

	logger := log.New(os.Stdout, "", log.Ldate+log.Ltime+log.Lshortfile)
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
