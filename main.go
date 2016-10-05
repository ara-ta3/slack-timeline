package main

import (
	"flag"
	"log"

	"./timeline"
)

func main() {
	filePath := flag.String("c", "config.json", "file path to config.json")
	flag.Parse()
	config, e := ReadConfig(filePath)
	if e != nil {
		log.Fatalln(e)
	}

	service := timeline.NewTimelineService(
		config.SlackAPIToken,
		config.TimelineChannelID,
		config.BlackListChannelIDs,
	)
	e := service.Run()
	if e != nil {
		log.Fatalf("%+v", e)
	}
}
