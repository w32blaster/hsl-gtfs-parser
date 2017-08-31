package main

import (
	"database/sql"
	"fmt"

	"github.com/patrickbr/gtfsparser"
)

const (
	dbFile      = "./db/helsinki_timetables.sqlite"
	archiveFile = "HSL.zip"
)

func main() {

	fmt.Println("Let's start")

	var db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// create Schemes
	CreateSchema(db)

	feed := gtfsparser.NewFeed()
	if err := feed.Parse(archiveFile); err == nil {

		fmt.Printf("Done, parsed %d agencies, %d stops, %d routes, %d trips, %d fare attributes\n\n",
			len(feed.Agencies), len(feed.Stops), len(feed.Routes), len(feed.Trips), len(feed.FareAttributes))

		// 1) insert cities
		InsertCities(feed, db)

		// 2) insert companies
		InsertCompanies(feed, db)

	} else {
		fmt.Println("Error in reading")
		fmt.Println(err)
	}

}
