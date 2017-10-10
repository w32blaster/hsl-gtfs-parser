package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	sets "github.com/deckarep/golang-set"
	_ "github.com/mattn/go-sqlite3"
	parser "github.com/patrickbr/gtfsparser"
	"github.com/patrickbr/gtfsparser/gtfs"
)

// CreateSchema : creates schema
func CreateSchema(db *sql.DB) bool {

	// Valid SQL to test:
	// SELECT * FROM transport_mode as tm INNER JOIN transport_number as tn ON tm._id = tn.transport_mode_id WHERE tm.vehicle_type_id = 1 GROUP BY tm._id ORDER BY tm._id, tn.name;

	sqlStmt := `
	DROP TABLE IF EXISTS city;
	DROP TABLE IF EXISTS station;
	DROP TABLE IF EXISTS point;
	DROP TABLE IF EXISTS transport_number;
	DROP TABLE IF EXISTS transport_mode;
	DROP TABLE IF EXISTS vehicle_type;
	DROP TABLE IF EXISTS company;
	DROP TABLE IF EXISTS trip;

	CREATE TABLE city (_id integer primary key, name VARCHAR(30) not null);
	CREATE TABLE station (_id integer primary key, name VARCHAR(50) not null,city_id integer not null);
	CREATE TABLE point (trip_id NUMERIC not null, station_id integer not null, time integer not null, idx integer not null);
	CREATE TABLE transport_number (_id integer primary key, name VARCHAR(30) not null,service_name VARCHAR(30) not null, transport_mode_id integer not null);
	CREATE TABLE transport_mode (_id integer primary key autoincrement, name VARCHAR(50) not null,vehicle_type_id integer not null);
	CREATE TABLE vehicle_type (_id integer primary key, name VARCHAR(20) not null);
	CREATE TABLE company (_id integer primary key, name VARCHAR(50) not null);
	CREATE TABLE trip (_id integer primary key, company_id integer not null,station_id_start integer not null,station_id_end integer not null,is_workday BOOLEAN not null,is_saturday BOOLEAN not null,is_sunday BOOLEAN not null,transport_number_id integer not null);
	
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

	_, err := db.Exec("insert into vehicle_type(_id, name) values(1, 'Bus'), (2, 'Trains'), (3, 'Metro'), (4, 'Trams'), (5, 'Ferry'), (6, 'U-Lines'), (7, 'Cable Cars'), (8, 'Gondola'), (9, 'Funicular')")
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

// InsertTransportModes : insert cities to the database
func InsertTransportModes(db *sql.DB) bool {
	_, err := db.Exec("insert into transport_mode (_id, name, vehicle_type_id) values(1, 'Bussiliikenne', 1), (2, 'Raitiovaunuliikenne', 4), (6, 'Metroliikenne', 3), (7, 'Vesiliikenne', 5), (8, 'U-liikenne', 6), (12, 'Lähijunaliikenne', 2), (21, 'Lähibussiliikenne', 1)")
	if err != nil {
		log.Fatal(err)
		return false
	}

	return false
}

// InsertTransportNumbers : insert cities to the database
func InsertTransportNumbers(feed *parser.Feed, db *sql.DB) *map[string]int {

	mapRouteIds := make(map[string]int)

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	id := 0
	stmt, err := tx.Prepare("insert into transport_number (_id, name, service_name, transport_mode_id) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for routeKey := range feed.Routes {

		_, err = stmt.Exec(id, feed.Routes[routeKey].Short_name, feed.Routes[routeKey].Long_name, getTransportType(feed.Routes[routeKey].Type))
		if err != nil {
			log.Fatal(err)
		}
		mapRouteIds[feed.Routes[routeKey].Id] = id
		id++
	}
	tx.Commit()

	fmt.Println("   inserted " + strconv.Itoa(len(feed.Routes)) + " transport numbers")
	return &mapRouteIds
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

	fmt.Println("   inserted " + strconv.Itoa(len(feed.Stops)) + " stations")
	return true
}

// InsertTripsAndPoints : insert cities to the database
func InsertTripsAndPoints(feed *parser.Feed, db *sql.DB, mapRouteIds *map[string]int) bool {

	hasSet := sets.NewThreadUnsafeSet()

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

	tripID := 0
	tripCount := len(feed.Trips)
	dublicatesCnt := 0

	uniqueSnapshot := ""

	for tripKey := range feed.Trips {

		if isNotDublicate(&feed.Trips[tripKey].Service.Daymap) {

			// get the first and the last station in the current trip
			stationStart := feed.Trips[tripKey].StopTimes[0].Stop.Id
			stopsCnt := len(feed.Trips[tripKey].StopTimes)
			stationEnd := feed.Trips[tripKey].StopTimes[stopsCnt-1].Stop.Id

			// find the dates (workdays, Saturday, Sunday)
			isMonday := isWorkday(&feed.Trips[tripKey].Service.Daymap)
			isSaturday := 0
			if feed.Trips[tripKey].Service.Daymap[6] {
				isSaturday = 1
			}

			isSunday := 0
			if feed.Trips[tripKey].Service.Daymap[0] {
				isSunday = 1
			}

			// before that we need to check with the following SQL:
			// SELECT count(*) as cnt FROM trip AS t INNER JOIN point AS p ON p.trip_id = t._id WHERE station_id_start=1520702 AND is_workday=1 AND is_saturday=0 AND is_sunday=0 AND p.time=0600;
			firstStationTime := extractTime(&feed.Trips[tripKey].StopTimes[0].Arrival_time)
			uniqueSnapshot = stationStart + "_" + stationEnd + "_" + strconv.Itoa(isMonday) + "_" + strconv.Itoa(isSaturday) + "_" + strconv.Itoa(isSunday) + "_" + strconv.Itoa(firstStationTime)

			if !hasSet.Contains(uniqueSnapshot) {

				// insert one trip
				stmt.Exec(tripID, 1, stationStart, stationEnd, isMonday, isSaturday, isSunday, (*mapRouteIds)[feed.Trips[tripKey].Route.Id])

				// iterate over all stops and insert them
				insertPoints(tx, feed.Trips[tripKey].StopTimes, tripID)

				if tripID%10 == 0 {
					fmt.Printf("\r Inserted %d/%d (dublicates: %d)", tripID, tripCount, dublicatesCnt)
				}

				hasSet.Add(uniqueSnapshot)
				tripID++
			} else {
				dublicatesCnt++
			}
		}
	}

	tx.Commit()
	fmt.Println("   inserted " + strconv.Itoa(tripID) + " trips")
	return true
}

// insert all the stop times (or, "points") for the given route
func insertPoints(tx *sql.Tx, stopTimes gtfs.StopTimes, tripID int) {

	//CREATE TABLE point (trip_id NUMERIC not null, station_id integer not null, time integer not null, idx integer not null);
	stmt, err := tx.Prepare("INSERT INTO point (trip_id, station_id, time, idx) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// iterate over all the stops and write them to the database
	for _, stopTime := range stopTimes {
		stmt.Exec(tripID, stopTime.Stop.Id, extractTime(&stopTime.Arrival_time), stopTime.Sequence)
	}
}

// import contains record for all the days, but we combine Mo, Tu, We, Th and Fr as "workdays",
// that's why each trip is dublicated for workdays 5 times. The simplest solution
// would be to take only recored for Monday assuming that for the Tu,We,Th and Fr timetables are the same.
// This method is used in order to take records only for Mon, Sun and Sat and ignore the rest of the days.
func isNotDublicate(dayMap *[7]bool) bool {

	// TRUE only if dayMap conains MON, SUN or SAT
	// indexes could be find here: https://github.com/geops/gtfsparser/blob/master/mapping.go#L93
	return dayMap[1] || dayMap[6] || dayMap[0]
}

// Shortcut method that returns 1 only when a trip is active in Wordays, 0 if not
// please, refer to the mapping example here: https://github.com/geops/gtfsparser/blob/master/mapping.go#L93
func isWorkday(dayMap *[7]bool) int {
	if dayMap[1] || dayMap[2] || dayMap[3] || dayMap[4] || dayMap[5] {
		return 1
	}
	return 0
}

// in order to simplify the calculations, we
// store time as ineger. So, 16:30 will be 1630, the 18:15 will be 1815.
// This allows to draw timetables faster.
// the current method takes the time and returns int that we expect to see
// please refer to the unit tests
func extractTime(time *gtfs.Time) int {
	return (int(time.Hour) * 100) + int(time.Minute)
}

// getTransportType maps the GTFS types to the old HSL types.
// Helsinki Timetables uses old types for trasport which are less granular, than new onces. So we analyze given type
// and change it to the expected.
// All the GTFS types can be found here: https://developers.google.com/transit/gtfs/reference/#routestxt
// (see route_type field) and extended types can be found here: https://developers.google.com/transit/gtfs/reference/extended-route-types
func getTransportType(gtfsType int16) int16 {

	/*
		old types are

			1, 'Bussiliikenne'
			2, 'Raitiovaunuliikenne'
			6, 'Metroliikenne'
			7, 'Vesiliikenne'
			8, 'U-liikenne'
			12, 'Lähijunaliikenne'
			21, 'Lähibussiliikenne'

	*/

	// standard
	switch gtfsType {
	case 0:
		return 2
	case 1:
		return 6
	case 2:
		return 12
	case 3:
		return 1
	case 4:
		return 7
	case 300:
		return 12
	case 500:
		return 6
	case 600:
		return 6
	case 1200:
		return 7
	}

	// extended
	switch {
	case (gtfsType >= 100 && gtfsType < 200):
		return 1
	case (gtfsType >= 200 && gtfsType < 300):
		return 21
	case (gtfsType >= 400 && gtfsType < 500):
		return 12
	case (gtfsType >= 700 && gtfsType < 800):
		return 1
	case (gtfsType >= 800 && gtfsType < 900):
		return 2
	case (gtfsType >= 1000 && gtfsType < 1100):
		return 2
	}

	return 1
}
