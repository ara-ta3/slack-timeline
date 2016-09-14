package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"./timeline"
)

func main() {
	blacklists := flag.String("blacklists", "", "comma separated black list channel ids")
	flag.Parse()

	service := timeline.NewTimelineService(
		os.Getenv("SLACK_TOKEN"),
		os.Getenv("SLACK_TIMELINE_CHANNEL_ID"),
		strings.Split(*blacklists, ","),
	)
	e := service.Run()
	if e != nil {
		log.Fatalf("%+v", e)
	}
}
