package mqttGather

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/a2800276/gocart/encoding/pem"
	myx509 "github.com/a2800276/gocart/x509"
)

// Functionality in this file is intended to provide minimal
// CA functionality for OpenAir/OpenNoise devices and backends.

type Bits int

const (
	B1024 = Bits(1024)
	B2048 = Bits(2048)
	B4096 = Bits(4096)
)

// The returned `rsa.PublicKey` and `crypto.PublicKey` are identical keys and
// provided as a utility to avoid casting.
func generateRSAKey(bits Bits) (*rsa.PrivateKey, *rsa.PublicKey, crypto.PublicKey) {
	priv, err := rsa.GenerateKey(rand.Reader, int(bits))
	if err != nil {
		panic(err)
	}

	if rsaPublicKey, ok := priv.Public().(*rsa.PublicKey); !ok {
		panic("generate RSA Key did not generate an RSA key")
	} else {
		return priv, rsaPublicKey, priv.Public()
	}

}

// template to create a certificate filename from a base filename, e.g:
//     "my_pki" => "my_pki.crt.pem"
func CertFn(fnBase string) string {
	return fmt.Sprintf("%s.crt.pem", fnBase)
}

// template to create a key filename from a base filename, e.g:
//     "my_pki" => "my_pki.key.pem"
func PrivKeyFn(fnBase string) string {

	return fmt.Sprintf("%s.key.pem", fnBase)
}

// Base certificate template for all certs.  this file provides functionality
// to generate CA, host and client certs and a similar template function is
// provided for each type of certificate.
func createBaseTemplate(pubKey *rsa.PublicKey) *myx509.CertificateData {
	data := new(myx509.CertificateData)
	data.NotBefore = time.Now()
	data.NotAfter = data.NotBefore.AddDate(5, 0, 0)

	data.BasicConstraintsValid = true

	return data

}

// Template for CA certs. see @createBaseTemplate and @CreateCACertificateAndKey
func createCATemplate(cn string, pubKey *rsa.PublicKey) *x509.Certificate {
	data := createBaseTemplate(pubKey)
	data.SerialNumber = myx509.GeneratePubKeyHash(pubKey)
	data.Subject = myx509.Subject{
		CommonName: cn,
	}

	data.KeyUsage = myx509.KeyUsage{
		CertSign: true,
	}

	data.IsCA = true
	if cert, err := data.CreateX509Template(); err != nil {
		panic(err)
	} else {
		return cert
	}
}

// Create a CA/signing certificate for Subject CN set to `cn`. The generated
// certificate and corresponding private key are store in (unencrypted) pem
// files named according to `fnBase`:
//
//     <fnBase>.crt.pem     and
//     <fnBase>.key.pem
//
// The CA cert used for signing is indicated by the `caFn` parameter which uses
// the same conventions.
//
// A 2048bit RSA keypair is generated, keyusage is set to "cert sign",
// `basic_constraints_value` set to `true` as is `is_ca`

func CreateCACertificateAndKey(cn string, fnBase string, bits Bits) error {
	priv, rsaPub, pub := generateRSAKey(bits)
	template := createCATemplate(cn, rsaPub)

	bytes, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		return err
	}

	privK_bytes, err := x509.MarshalPKCS8PrivateKey(priv)

	err = pem.StorePEM(PrivKeyFn(fnBase), pem.PEM_PRIVATE_KEY, privK_bytes)
	if err != nil {
		return err
	}

	return pem.StorePEM(CertFn(fnBase), pem.PEM_CERTIFICATE, bytes)
}

// Template for CA certs. see @createBaseTemplate and @CreateClientCertificateAndKey
func createClientTemplate(mac string, pubKey *rsa.PublicKey) (*x509.Certificate, error) {
	data := createBaseTemplate(pubKey)
	mac_, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}
	data.SerialNumber = myx509.NewHex([]byte(mac_))
	data.Subject = myx509.Subject{
		CommonName: mac,
	}

	data.KeyUsage = myx509.KeyUsage{
		DigitalSignature: true,
	}

	if cert, err := data.CreateX509Template(); err != nil {
		panic(err)
	} else {
		return cert, nil
	}
}

// Create a client certificate for the specified `macAddr`. The generated
// certificate and corresponding private key are store in (unencrypted) pem
// files named according to `baseFn`:
//
//     <baseFn>.crt.pem     and
//     <baseFn>.key.pem
//
// The CA cert used for signing is indicated by the `caFn` parameter which uses
// the same conventions.

// A 2048bit RSA keypair and more or less generic certificate intendende for
// client authentication against a mosquitto mqqt server. mosquitto require
// 2048 bit minimum RSA keys  (which may not be optimal for IOT devices ...)
// The provided MAC address is used as both the certificate serial number and
// CN field in the certificates Subject. Mosquitto can be configured to
// recognize the CN as username/client_id using the `user_identity_as_username`
// configuration parameter. (see [mosquitto
// conf](https://mosquitto.org/man/mosquitto-conf-5.html))

func CreateClientCertificateAndKey(macAddr string, baseFn string, caFn string) error {
	priv, rsaPub, pub := generateRSAKey(B2048)
	template, err := createClientTemplate(macAddr, rsaPub)
	if err != nil {
		return err
	}

	ca_cert, err := pem.LoadPEMCertificate(CertFn(caFn))
	if err != nil {
		return err
	}

	ca_key, err := pem.LoadPEMFile(PrivKeyFn(caFn), nil)
	if err != nil {
		return err
	}

	bytes, err := x509.CreateCertificate(rand.Reader, template, ca_cert, pub, ca_key)

	if err != nil {
		return err
	}

	privK_bytes, err := x509.MarshalPKCS8PrivateKey(priv)

	err = pem.StorePEM(PrivKeyFn(baseFn), pem.PEM_PRIVATE_KEY, privK_bytes)
	if err != nil {
		return err
	}

	return pem.StorePEM(CertFn(baseFn), pem.PEM_CERTIFICATE, bytes)
}

// paramter for the host certificate, see @CreateMQTTServerCertificateAndKey
func createHostTemplate(hostname string, pubKey *rsa.PublicKey) (*x509.Certificate, error) {
	data := createBaseTemplate(pubKey)
	data.SerialNumber = myx509.GeneratePubKeyHash(pubKey)
	data.Subject = myx509.Subject{
		CommonName: hostname,
	}

	if cert, err := data.CreateX509Template(); err != nil {
		panic(err)
	} else {
		return cert, nil
	}
}

// Create a Server certificate for the specified `hostname`. The generated
// certificate and corresponding private key are store in (unencrypted) pem
// files named according to `fnBase`:
//
//     <fnBase>.crt.pem     and
//     <fnBase>.key.pem
//
// The CA cert used for signing is indicated by the `caFn` parameter which uses
// the same conventions.
//
// A 2048bit RSA keypair and more or less generic certificate with hostname in
// the Subject CN field is generated.

func CreateMQTTServerCertificateAndKey(hostname string, fnBase string, caFn string) error {
	priv, rsaPub, pub := generateRSAKey(B2048)
	template, err := createHostTemplate(hostname, rsaPub)
	if err != nil {
		return err
	}

	ca_cert, err := pem.LoadPEMCertificate(CertFn(caFn))
	if err != nil {
		return err
	}

	ca_key, err := pem.LoadPEMFile(PrivKeyFn(caFn), nil)
	if err != nil {
		return err
	}

	bytes, err := x509.CreateCertificate(rand.Reader, template, ca_cert, pub, ca_key)

	if err != nil {
		return err
	}

	privK_bytes, err := x509.MarshalPKCS8PrivateKey(priv)

	err = pem.StorePEM(PrivKeyFn(fnBase), pem.PEM_PRIVATE_KEY, privK_bytes)
	if err != nil {
		return err
	}

	return pem.StorePEM(CertFn(fnBase), pem.PEM_CERTIFICATE, bytes)

}
