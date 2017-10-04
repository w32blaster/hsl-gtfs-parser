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
		fmt.Println("1) Insert sities")
		InsertCities(db)

		// 2) insert companies
		fmt.Println("2) insert companies")
		InsertCompanies(feed, db)

		// 3) insert transport types
		fmt.Println("3) insert transport types")
		InsertVehicleTypes(db)

		// 4) insert all the transport Numbers and route names
		fmt.Println("4) instert transport numbers")
		InsertTransportNumbers(feed, db)

		// 5) stations
		fmt.Println("5) insert stations")
		InsertStations(feed, db)

		// 6) points
		fmt.Println("6) insert trips points (bus stops)")
		InsertTripsAndPoints(feed, db)

		fmt.Println("Done")

	} else {
		fmt.Println("Error in reading")
		fmt.Println(err)
	}

}
