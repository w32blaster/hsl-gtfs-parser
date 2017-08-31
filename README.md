# HSL dump parser

The Helsinki Region Transport department ([HSL](https://www.hsl.fi/)) exports all the data for public transport for Helsinki city, Finland in the GTFS format.

You can find the dump data [on the Reittiopas site](http://developer.reittiopas.fi/pages/en/other-apis.php). Please refer also to the [official GTFS documentation](https://developers.google.com/transit/gtfs/) for details about the format used.

This utility **HSL dump parser** parses data from Reittiopas and generates ready to use SQlite database. It can be used by any application, but I use it in the [Helsinki Timetables](https://github.com/w32blaster/helsinki-timetables) Android application.

Used libraries:
* https://github.com/patrickbr/gtfsparser
* https://github.com/mattn/go-sqlite3