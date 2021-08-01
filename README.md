
# Functionality to collect MQTT data into DB

	Usage of release/mqttGather.v0.2.1.linux:
	  -c string
		name of (optional) config file
	  -clientID string
		clientId to use for connection
	  -host string
		host to connect to
	  -silent
		psssh!
	  -sqlite string
		connect string to use for sqlite, when in doubt: provide a filename
	  -topic string
		topic to subscribe to
	  -telemetry-topic string
	        telemetry topic to subscribe to
	  -version
		display version information and exit

Flags should hopefully be obvious. The `-silent` flag currently only
supresses the initial banner providing version and connection info.


## Config File

In order to avoid having to type out a a bunch of flags all the time,
they may be collected in a JSON config as follows:

	{
		"sqlite":":memory:",
		"host":"tcp://test.mosquitto.org:1883",
		"topic":"/opennoise/+/dba_stats"
		"telemetry-topic":"/opennoise/+/telemetry",
		"client_id":"mqttTest"
	}

Flags provided on the command line take priority over those in the
config file.

## Building

Source the `xcompile.sh` script which builds executables for linux,
windows and macos. This also extracts the last accessible tag from the
git repository, which is hopefully the release we are trying to build,
using:

	git describe --tags --dirty

(In case we are further along than the tag, `-dirty` is appended)

The complete binaries are located in the folder `release` with names appended
to contain version and plattform information:

	$ ls release
	mqttGather.v0.1.darwin
	mqttGather.v0.1.linux
	mqttGather.v0.1.windows
	mqttGather.v0.2.1RC1.darwin
	mqttGather.v0.2.1RC1.linux
	mqttGather.v0.2.1RC1.windows

Compiled release binaries may also be downloaded from the Github [release
page](https://github.com/OpenAirCgn/mqttGather/releases)


## TODOS
- telemetry: handle flag and ESQ values
- db : denormalize client and migrate
- IN PROGRESS Weather Data Import: https://www.dwd.de/DE/leistungen/klimadatendeutschland/klimadatendeutschland.html
- -silent should suppress logging
- logging to file / DB with rotation
- TLS
- different backends
- different plugins/topics to gather other sensor data
- generate random / uuid / mac based clientids to not kick other clients
