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

	sqlStmt := `
	DROP TABLE city;
	DROP TABLE station;
	DROP TABLE point;
	DROP TABLE transport_number;
	DROP TABLE transport_mode;
	DROP TABLE vehicle_type;
	DROP TABLE company;
	DROP TABLE trip;

	CREATE TABLE city (_id integer primary key autoincrement, name VARCHAR(30) not null);
	CREATE TABLE station (_id integer primary key autoincrement, name VARCHAR(50) not null,city_id integer not null);
	CREATE TABLE point (trip_id NUMERIC not null, station_id integer not null, time integer not null, idx integer not null);
	CREATE TABLE transport_number (_id integer primary key autoincrement, name VARCHAR(30) not null,service_name VARCHAR(30) not null,transport_mode_id integer not null);
	CREATE TABLE transport_mode (_id integer primary key autoincrement, name VARCHAR(50) not null,vehicle_type_id integer not null);
	CREATE TABLE vehicle_type (_id integer primary key autoincrement, name VARCHAR(20) not null);
	CREATE TABLE company (_id integer primary key autoincrement, name VARCHAR(50) not null);
	CREATE TABLE trip (_id integer primary key autoincrement, company_id integer not null,station_id_start integer not null,station_id_end integer not null,is_workday BOOLEAN not null,is_saturday BOOLEAN not null,is_sunday BOOLEAN not null,transport_number_id integer not null);
	
	CREATE INDEX statNameIdx ON station(name);
	CREATE INDEX idxPointTrip ON point( trip_id);
	CREATE INDEX idxPointStat ON point(station_id);
	CREATE INDEX idx4 ON transport_number(transport_mode_id);
	CREATE INDEX idx5 ON transport_mode(vehicle_type_id);
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
func InsertCities(feed *parser.Feed, db *sql.DB) bool {

	_, err := db.Exec("insert into city(_id, name) values(1, 'Helsinki'), (2, 'Espoo'), (3, 'Kauniainen'), (4, 'Vantaa'), (6, 'Kirkkonummi'), (9, 'Kerava')")
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

// InsertCompanies : insert cities to the database
func InsertCompanies(feed *parser.Feed, db *sql.DB) bool {
	fmt.Print(feed.Agencies["HSL"])

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

// InsertVehicleTypes : insert cities to the database
func InsertVehicleTypes() bool {
	return false
}

// InsertTransportModes : insert cities to the database
func InsertTransportModes() bool {
	return false
}

// InsertTransportNumbers : insert cities to the database
func InsertTransportNumbers() bool {
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
