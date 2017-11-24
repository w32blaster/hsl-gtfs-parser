# HSL dump parser

The Helsinki Region Transport department ([HSL](https://www.hsl.fi/)) exports all the data for public transport for Helsinki city, Finland in the GTFS format.

The official Docker Image is [w32blaster/hsl-gtfs-parser](https://hub.docker.com/r/w32blaster/hsl-gtfs-parser/).

You can find the dump data [on the Reittiopas site](http://developer.reittiopas.fi/pages/en/other-apis.php). Please refer also to the [official GTFS documentation](https://developers.google.com/transit/gtfs/) for details about the format used.

This utility **HSL dump parser** parses data from Reittiopas and generates ready to use SQlite database. It can be used by any application, but I use it in the [Helsinki Timetables](https://github.com/w32blaster/helsinki-timetables) Android application.

To install all the dependencies to your vendor folder, please use the command:

'''
$ govendor fetch +out
'''

Then you can build your app as usually. Make sure you have downkloaded the HSL.zip file using the link: https://api.digitransit.fi/routing-data/v2/hsl/HSL.zip

# Build a container

Please note, we use ["multi-stage build"](https://docs.docker.com/engine/userguide/eng-image/multistage-build/), that requires the latest version of Docker-ce. To build, use `docker build .`
