
# Functionality to collect MQTT data into DB

	Usage of /tmp/go-build892785999/b001/exe/main:
	  -c string
		name of (optional) config file
	  -clientID string
		clientId to use for connection
	  -host string
		host to connect to
	  -log-dir string
		where to write logs, writes to stdout if not set
	  -silent
		psssh!
	  -sms-key string
		api key for SMS
	  -sqlite string
		connect string to use for sqlite, when in doubt: provide a filename
	  -telemetry-topic string
		topic to subscribe to for telemetry data
	  -topic string
		topic to subscribe to
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
		"client_id":"mqttTest",
		"sms_key":"asdfasfds"
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

The complete binaries are located in the folder `release`:

	$ ls release
	mqttGather.v0.1.darwin
	mqttGather.v0.1.linux
	mqttGather.v0.1.windows
	mqttGather.v0.2.1RC1.darwin
	mqttGather.v0.2.1RC1.linux
	mqttGather.v0.2.1RC1.windows

Currently, the binaries are not crosscompiled, because sqlite driver
contains native code which makes crosscompiled binaries more difficult
to create. 

Compiled release binaries may also be downloaded from the Github [release
page](https://github.com/OpenAirCgn/mqttGather/releases)

## Testing

During tests, SMS alerts are typically NOT sent to avoid being annoying.
You can set an environment variable named SMSKEY to actually send an SMS.

## VERSIONS

### 0.4.0 

- added alert functionality
- added log rotation

## TODOS
- telemetry: handle flag and ESQ values
- db : denormalize client and migrate
- IN PROGRESS Weather Data Import: https://www.dwd.de/DE/leistungen/klimadatendeutschland/klimadatendeutschland.html
- -silent should suppress logging
- TLS
- different backends
- different plugins/topics to gather other sensor data
- generate random / uuid / mac based clientids to not kick other clients
