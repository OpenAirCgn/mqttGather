# Overview

This project currently contains two utilities:

- mqttGather : an mqtt Subscriber meant to gather OpenAir data
  transmitted from devices to an mqtt broker and serialize the
  measurements to a database.

- ca : a mini certificate authority to quickly generate the necessary
  TLS certificates for devices and servers.

# Building

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
	<mqttGather|ca>.<git version>.<plattform>

Compiled release binaries may also be downloaded from the Github [release
page](https://github.com/OpenAirCgn/mqttGather/releases)


# `mqttGather` : Functionality to collect MQTT data into DB

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

# `ca` : mini Certificate Authority 

This utility may be used to create (primarily) client certificates for
the OpenAir IoT devices, geared towards use in the OpenAirCgn
infrastructure. As a consequence, this is not a general purpose CA, e.g.
all certificates have underlying 2048-bit keys, IoT client certs have
the devices MAC as serialnumber and Subject CN, the only config
available to server certs is the hostname, which will be set as CN. All
certificates are valid for five years from their creation. (This should
be all for now, as we add more these points should be move to their own
section)

## Usage

	Usage of ./ca:
	  -caCommonName string
		CN to use in root CA certificate (default "www.example.com")
	  -caFilename (...).crt.pem
		basename to use when storing CA cert and key in pems named (...).crt.pem and `(...).key.pem`, defaults to `-caCN`
	  -certDir string
		directory to store client certs (default ".")
	  -deviceList string
		file containing list of device macs
	  -deviceMAC -deviceList
		MAC to use for certificate serial number and CN, use -deviceList if creating multiple certs (default "12:23:34:45:56:67")
	  -force
		use to override existing crts and keys
	  -generateAll
		implies -generateCA -generateHost and -generateDevice
	  -generateCA
		generate a CA certificate and key
	  -generateDevice
		generate device key(s)
	  -generateHost
		generate a host/server certificate
	  -hostFilename <-hostname>.crt.pem
		basename to use storing host cert & key. Defaults to <-hostname>.crt.pem and `.key.pem`
	  -hostname string
		hostname to use in server certificate's CN (default "localhost")
	  -version
		display version information and exit

## Initial Setup

In case you will be self-signing certificates (the expected use case)
you will need to create a root ca certificate:

	$ ca -generateCA -caCommonName "Whatever you like" -caFilename certs/myCA

This will result in the following to files, containing certificate and
key being stored in:

	- certs/myCA.crt.pem
	- certs/myCA.key.pem

Note that the key file is not encrypted!

## Generating and Signing Certificates

Because the intended use case is to provide self-signed certificates for
internal use, we can generate and immediately sign certificates. There
is no need for CSRs.

To generate a host certificate, e.g. for an MQTT Server:

	$ ca -generateHost -hostname www.example.com -caFilename certs/myCA

generates the files:

	- www.example.com.crt.pem
	- www.example.com.key.pem

in the current directory. The certificate/key:
`certs/myCa.(crt|key).pem` generated by the previous example are used to
sign the certificate.

### Client Certificates for Devices

You can provide a file containing a list of MAC addresses (one per line)
and pass them using the `-deviceList` flag, e.g.

	$ ca -deviceList <file containing list of MACs> -generateDevice -caFilename certs/myCA



## mosquitto configuration

Assuming you are using a mosquitto instance to authenticate against, the
following configuration options are relevant and need to be set. Please
refer to mosquitto-conf documentation for further details:


	# For the sake of this example we are configuring mosquitto with a 
	# single default listener for TLS on port 8883

	port 8883

	# The CA file is used by mosquitto to verify the authenticity of
	# incoming client certificates
	cafile /full/path/to/certs/myCA.crt.pem

	# `certfile` and `keyfile` are the basic files for the TLS listener,
	# i.e. the host certicate
	certfile /full/path/to/www.example.com.crt.pem
	keyfile  /full/path/to/www.example.com.key.pem

	# `require_certificate` tells mosquitto to definitely use cert based
	# authentication
	require_certificate true

	# .. and finally `use_identity_as_username` indicates that the contents
	# of the Common Name / CN element of the certificate's subject is to be
	# considered the username / client_id
	use_identity_as_username true


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
