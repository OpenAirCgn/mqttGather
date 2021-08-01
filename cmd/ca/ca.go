package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/openaircgn/mqttGather"
)

var (
	all = flag.Bool("generateAll", false, "implies -generateCA -generateHost and -generateDevice")

	ca   = flag.Bool("generateCA", false, "generate a CA certificate and key")
	caCN = flag.String("caCommonName", "www.example.com", "CN to use in root CA certificate")
	caFn = flag.String("caFilename", "", "basename to use when storing CA cert and key in pems named `(...).crt.pem` and `(...).key.pem`, defaults to `-caCN`")

	host       = flag.Bool("generateHost", false, "generate a host/server certificate")
	hostname   = flag.String("hostname", "localhost", "hostname to use in server certificate's CN")
	hostnameFn = flag.String("hostFilename", "", "basename to use storing host cert & key. Defaults to `<-hostname>.crt.pem` and `.key.pem`")

	device     = flag.Bool("generateDevice", false, "generate device key(s)")
	deviceMac  = flag.String("deviceMAC", "12:23:34:45:56:67", "MAC to use for certificate serial number and CN, use `-deviceList` if creating multiple certs")
	deviceList = flag.String("deviceList", "", "file containing list of device macs")

	certDir = flag.String("certDir", ".", "directory to store client certs")

	force = flag.Bool("force", false, "use to override existing crts and keys")

	_version = flag.Bool("version", false, "display version information and exit")
)

var version string /* filled by linker during build*/

func banner() {
	fmt.Fprintf(os.Stderr, "%s ver %s\n", os.Args[0], version)
}

func fileOfLinesToArray(fn string) ([]string, error) {
	file, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func exists(baseFn string) bool {
	check := func(fn string) bool {
		_, err := os.Stat(fn)
		if err != nil {
			if os.IsNotExist(err) {
				return false
			}
			panic(err)
		}
		return true
	}
	return check(mqttGather.CertFn(baseFn)) && check(mqttGather.PrivKeyFn(baseFn))

}
func usage(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	flag.Usage()
	os.Exit(1)
}

func main() {

	flag.Parse()

	banner()
	if *_version {
		os.Exit(0)
	}

	if *all {
		*ca = true
		*host = true
		*device = true
	}

	if *caFn == "" {
		*caFn = *caCN
	}

	if *ca {
		if exists(*caFn) && !*force {
			usage("output CA file(s) already exists. Cowardly giving up. Use -force to overwrite")
		}

		fmt.Fprintf(os.Stderr, "Creating CA cert and key for: %s writing to: %s.(crt|key).pem\n", *caCN, *caFn)
		if err := mqttGather.CreateCACertificateAndKey(*caCN, *caFn, mqttGather.B2048); err != nil {
			panic(err)
		}
	}

	if *host {
		if *hostnameFn == "" {
			*hostnameFn = *hostname
		}

		if exists(*hostnameFn) && !*force {
			usage("output host file(s) already exists. Cowardly giving up. Use -force to overwrite")
		}
		fmt.Fprintf(os.Stderr, "Creating Host certificate for: %s writing to: %s.(crt|key).pem\n", *hostname, *hostnameFn)
		fmt.Fprintf(os.Stderr, "    using: %s as CA cert for signing\n", *caFn)

		if err := mqttGather.CreateMQTTServerCertificateAndKey(*hostname, *hostnameFn, *caFn); err != nil {
			panic(err)
		}

	}

	if *device {
		var deviceMACs []string

		if *deviceList != "" {
			var err error
			deviceMACs, err = fileOfLinesToArray(*deviceList)
			if err != nil {
				panic(err)
			}
		} else {
			deviceMACs = append(deviceMACs, *deviceMac)
		}

		for _, mac := range deviceMACs {
			deviceFn := fmt.Sprintf("%s/%s", *certDir, mac)

			if exists(deviceFn) && !*force {
				fmt.Fprintf(os.Stderr, "Device files: %s exist, skipping. Use -force to overwrite", deviceFn)
				continue
			}

			fmt.Fprintf(os.Stderr, "Creating client certificate for: %s writing to: %s.(crt|key).pem\n", mac, deviceFn)
			fmt.Fprintf(os.Stderr, "    using: %s as CA cert for signing\n", *caFn)

			if err := mqttGather.CreateClientCertificateAndKey(mac, deviceFn, *caFn); err != nil {
				panic(err)
			}
		}
	}
}
