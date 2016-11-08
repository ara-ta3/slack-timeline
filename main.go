package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"

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

	logger := log.New(os.Stdout, "", log.Ldate+log.Ltime+log.Lshortfile)
	service := timeline.NewTimelineService(
		config.SlackAPIToken,
		config.TimelineChannelID,
		config.BlackListChannelIDs,
		*db,
		*logger,
	)
	err := service.Run()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
