package main

import (
	"log"
	"os"

	"./timeline"
)

func main() {
	service := timeline.NewTimelineService(os.Getenv("SLACK_TOKEN"), os.Getenv("SLACK_TIMELINE_CHANNEL_ID"))
	e := service.Run()
	if e != nil {
		log.Fatalf("%+v", e)
	}
}
