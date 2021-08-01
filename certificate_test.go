package mqttGather

import "testing"

func TestGenerateRSAKey(t *testing.T) {
	generateRSAKey(B1024)
	// this is to make sure the crypto.PublicKey / rsa.PublicKey
	// cast/type assertion works as expected
}
