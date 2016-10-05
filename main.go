package main

import (
	"flag"
	"log"
	"os"

	"./timeline"
)

func main() {
	filePath := flag.String("c", "config.json", "file path to config.json")
	flag.Parse()
	config, e := ReadConfig(*filePath)
	if e != nil {
		log.Fatalln(e)
	}

	logger := log.New(os.Stdout, "", log.Ldate+log.Ltime+log.Lshortfile)
	service := timeline.NewTimelineService(
		config.SlackAPIToken,
		config.TimelineChannelID,
		config.BlackListChannelIDs,
		*logger,
	)
	err := service.Run()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
