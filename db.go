package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	parser "github.com/patrickbr/gtfsparser"
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
	DROP TABLE city;
	DROP TABLE station;
	DROP TABLE point;
	DROP TABLE transport_number;
	DROP TABLE vehicle_type;
	DROP TABLE company;
	DROP TABLE trip;

	CREATE TABLE city (_id integer primary key autoincrement, name VARCHAR(30) not null);
	CREATE TABLE station (_id integer primary key autoincrement, name VARCHAR(50) not null,city_id integer not null);
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

	for k := range feed.Agencies {
		fmt.Print(" - " + k)
	}

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

	return false
}

// InsertStations : insert cities to the database
func InsertStations() bool {
	return false
}

// InsertTrips : insert cities to the database
func InsertTrips() bool {
	return false
}

// InsertPoints : insert cities to the database
func InsertPoints() bool {
	return false
}
