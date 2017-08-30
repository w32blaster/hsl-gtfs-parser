package main

import (
	"fmt"

	"github.com/geops/gtfsparser"
)

const (
	dbFile = "./db/trans.sqlite3"
)

func main() {

	fmt.Println("Let's start")

	feed := gtfsparser.NewFeed()
	if err := feed.Parse("HSL.zip"); err == nil {

		fmt.Printf("Done, parsed %d agencies, %d stops, %d routes, %d trips, %d fare attributes\n\n",
			len(feed.Agencies), len(feed.Stops), len(feed.Routes), len(feed.Trips), len(feed.FareAttributes))

	} else {
		fmt.Println("Error in reading")
		fmt.Println(err)
	}

}
