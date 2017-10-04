package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	parser "github.com/patrickbr/gtfsparser"
	"github.com/patrickbr/gtfsparser/gtfs"
)

// CreateSchema : creates schema
func CreateSchema(db *sql.DB) bool {

	// these lines are ignored:
	// DROP TABLE transport_mode;
	// CREATE TABLE transport_mode (_id integer primary key autoincrement, name VARCHAR(50) not null,vehicle_type_id integer not null);
	// CREATE INDEX idx4 ON transport_number(transport_mode_id);
	// CREATE INDEX idx5 ON transport_mode(vehicle_type_id);
	// (... ,transport_mode_id integer not null) from transport mode table

	sqlStmt := `
	DROP TABLE IF EXISTS city;
	DROP TABLE IF EXISTS station;
	DROP TABLE IF EXISTS point;
	DROP TABLE IF EXISTS transport_number;
	DROP TABLE IF EXISTS vehicle_type;
	DROP TABLE IF EXISTS company;
	DROP TABLE IF EXISTS trip;

	CREATE TABLE city (_id integer primary key autoincrement, name VARCHAR(30) not null);
	CREATE TABLE station (_id integer primary key, name VARCHAR(50) not null,city_id integer not null);
	CREATE TABLE point (trip_id NUMERIC not null, station_id integer not null, time integer not null, idx integer not null);
	CREATE TABLE transport_number (_id integer primary key autoincrement, name VARCHAR(30) not null,service_name VARCHAR(30) not null);
	CREATE TABLE vehicle_type (_id integer primary key autoincrement, name VARCHAR(20) not null);
	CREATE TABLE company (_id integer primary key autoincrement, name VARCHAR(50) not null);
	CREATE TABLE trip (_id integer primary key autoincrement, company_id integer not null,station_id_start integer not null,station_id_end integer not null,is_workday BOOLEAN not null,is_saturday BOOLEAN not null,is_sunday BOOLEAN not null,transport_number_id integer not null);
	
	CREATE INDEX statNameIdx ON station(name);
	CREATE INDEX idxPointTrip ON point( trip_id);
	CREATE INDEX idxPointStat ON point(station_id);
	CREATE INDEX idx1 ON trip ( transport_number_id);	
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return false
	}

	return true
}

// InsertCities : insert cities to the database
func InsertCities(db *sql.DB) bool {

	_, err := db.Exec("insert into city(_id, name) values(1, 'Helsinki'), (2, 'Espoo'), (3, 'Kauniainen'), (4, 'Vantaa'), (6, 'Kirkkonummi'), (9, 'Kerava')")
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

// InsertCompanies : insert cities to the database
func InsertCompanies(feed *parser.Feed, db *sql.DB) bool {

	_, err := db.Exec("insert into company(_id, name) values(1, '" + feed.Agencies["HSL"].Name + "')")
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

// InsertVehicleTypes : insert verhicle types. I found types in the docs
// https://developers.google.com/transit/gtfs/reference/routes-file#route_type
func InsertVehicleTypes(db *sql.DB) bool {

	_, err := db.Exec("insert into vehicle_type(_id, name) values(0, 'Tram'), (1, 'Metro'), (2, 'Trains'), (3, 'Bus'), (4, 'Ferry'), (5, 'Cable Cars'), (6, 'Gondola'), (7, 'Funicular')")
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

// InsertTransportModes : insert cities to the database
func InsertTransportModes() bool {
	// obsolete database, needs to rethink what to do with it :( may be remove
	return false
}

// InsertTransportNumbers : insert cities to the database
func InsertTransportNumbers(feed *parser.Feed, db *sql.DB) bool {

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into transport_number (name, service_name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer stmt.Close()

	for routeKey := range feed.Routes {
		_, err = stmt.Exec(feed.Routes[routeKey].Short_name, feed.Routes[routeKey].Long_name)
		if err != nil {
			log.Fatal(err)
			return false
		}
	}
	tx.Commit()

	return true
}

// InsertStations : insert cities to the database
func InsertStations(feed *parser.Feed, db *sql.DB) bool {

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
		return false
	}

	stmt, err := tx.Prepare("INSERT INTO station (_id, name, city_id) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer stmt.Close()

	// insert stops (stations) to the database
	for stopKey := range feed.Stops {
		// hopefully, "zone_id" matches the old "cities" ID
		stmt.Exec(feed.Stops[stopKey].Id, feed.Stops[stopKey].Name, feed.Stops[stopKey].Zone_id)
	}
	tx.Commit()

	return true
}

/*

Insert Poits and Trip:
1) load all the stop_times in a map, grouping by trip_id
2) insert first and last stop to TRIP
3) insert times to POINT

*/

// InsertTripsAndPoints : insert cities to the database
func InsertTripsAndPoints(feed *parser.Feed, db *sql.DB) bool {

	//CREATE TABLE trip (_id integer primary key autoincrement, company_id integer not null,station_id_start integer not null,station_id_end integer not null,is_workday BOOLEAN not null,is_saturday BOOLEAN not null,is_sunday BOOLEAN not null,transport_number_id integer not null);
	tx, err := db.Begin()

	if err != nil {
		log.Fatal(err)
		return false
	}

	stmt, err := tx.Prepare("INSERT INTO trip (_id, company_id, station_id_start, station_id_end, is_workday, is_saturday, is_sunday, transport_number_id) VALUES (?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer stmt.Close()

	id := 0
	for tripKey := range feed.Trips {

		// insert one trip
		stationStart := feed.Trips[tripKey].StopTimes[0].Stop.Id
		stopsCnt := len(feed.Trips[tripKey].StopTimes)
		stationEnd := feed.Trips[tripKey].StopTimes[stopsCnt-1].Stop.Id

		stmt.Exec(id, 1, stationStart, stationEnd, 0, 0, 0, -1)

		// iterate over all stops and insert them
		insertPoints(tx, feed.Trips[tripKey].StopTimes, id)

		id++
	}

	tx.Commit()
	return true
}

// insert all the stop times (or, "points") for the given route
func insertPoints(tx *sql.Tx, stops gtfs.StopTimes, tripID int) {

	stmt, err := tx.Prepare("INSERT INTO point (trip_id, station_id, time, idx) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// iterate over all the stops and write them to the database
	for _, stopTime := range stops {
		stmt.Exec(tripID, stopTime.Stop.Id, stopTime.Arrival_time, stopTime.Sequence)
	}
}
