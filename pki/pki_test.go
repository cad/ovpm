package pki_test

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/cad/ovpm/pki"
)

func TestNewCA(t *testing.T) {
	// Initialize:
	// Prepare:
	ca, err := pki.NewCA()
	if err != nil {
		t.Fatalf("can not create CA in test: %v", err)
	}

	// Test:
	// Is CertHolder empty?
	if ca.CertHolder == (pki.CertHolder{}) {
		t.Errorf("returned ca.CertHolder can't be empty: %+v", ca.CertHolder)
	}

	// Is CSR empty length?
	if len(ca.CSR) == 0 {
		t.Errorf("returned ca.CSR is a zero-length string")
	}

	var encodingtests = []struct {
		name  string // name
		block string // pem block string
		typ   string // expected pem block type
	}{
		{"ca.CSR", ca.CSR, pki.PEMCSRBlockType},
		{"ca.CertHolder.Cert", ca.CertHolder.Cert, pki.PEMCertificateBlockType},
		{"ca.CertHolder.Key", ca.CertHolder.Key, pki.PEMRSAPrivateKeyBlockType},
	}

	// Is PEM encoded properly?
	for _, tt := range encodingtests {
		if !isPEMEncodedProperly(t, tt.block, tt.typ) {
			t.Errorf("returned '%s' is not PEM encoded properly: %+v", tt.name, tt.block)
		}
	}

}

// TestNewCertHolders tests pki.NewServerCertHolder and pki.NewClientCertHolder functions.
func TestNewCertHolders(t *testing.T) {
	// Initialize:
	ca, _ := pki.NewCA()

	// Prepare:
	sch, err := pki.NewServerCertHolder(ca)
	if err != nil {
		t.Fatalf("can not create server cert holder: %v", err)
	}
	cch, err := pki.NewClientCertHolder(ca, "test-user")
	if err != nil {
		t.Fatalf("can not create client cert holder: %v", err)
	}

	// Test:
	var certholdertests = []struct {
		name       string
		certHolder *pki.CertHolder
	}{
		{"server", sch},
		{"client", cch},
	}

	for _, tt := range certholdertests {

		// Is CertHolder empty?
		if *tt.certHolder == (pki.CertHolder{}) {
			t.Errorf("returned '%s' cert holder can't be empty: %+v", tt.name, sch)
		}

		var encodingtests = []struct {
			name  string // name
			block string // pem block string
			typ   string // expected pem block type
		}{
			{tt.name + "CertHolder.Cert", tt.certHolder.Cert, pki.PEMCertificateBlockType},
			{tt.name + "CertHolder.Key", tt.certHolder.Key, pki.PEMRSAPrivateKeyBlockType},
		}

		// Is PEM encoded properly?
		for _, tt := range encodingtests {
			if !isPEMEncodedProperly(t, tt.block, tt.typ) {
				t.Errorf("returned '%s' is not PEM encoded properly: %+v", tt.name, tt.block)
			}
		}

	}

}

func TestNewCRL(t *testing.T) {
	// Initialize:
	max := 5
	n := randomBetween(1, max)
	ca, _ := pki.NewCA()

	// Prepare:
	var certHolders []*pki.CertHolder
	for i := 0; i < max; i++ {
		username := fmt.Sprintf("user-%d", i)
		ch, _ := pki.NewClientCertHolder(ca, username)
		certHolders = append(certHolders, ch)
	}

	// Test:
	// Create CRL that revokes first n certificates.
	var serials []*big.Int
	for i := 0; i < n; i++ {
		serials = append(serials, getSerial(t, certHolders[i].Cert))
	}

	crl, err := pki.NewCRL(ca, serials...)
	if err != nil {
		t.Fatalf("crl can not be created: %v", err)
	}

	// Is CRL empty?
	if len(crl) == 0 {
		t.Fatalf("CRL length expected to be NOT EMPTY %+v", crl)
	}

	// Is CRL PEM encoded properly?
	if !isPEMEncodedProperly(t, crl, pki.PEMx509CRLBlockType) {
		t.Fatalf("CRL is expected to be properly PEM encoded %+v", crl)
	}

	// Parse CRL and get revoked certList.
	block, _ := pem.Decode([]byte(crl))
	certList, err := x509.ParseCRL(block.Bytes)
	if err != nil {
		t.Fatalf("CRL's PEM block is expected to be parsed '%+v' but instead it CAN'T BE PARSED: %v", block, err)
	}

	rcl := certList.TBSCertList.RevokedCertificates

	// Is revoked cert list length is n, as correctly?
	if len(rcl) != n {
		t.Fatalf("revoked cert list lenth is expected to be %d but it is %d", n, len(rcl))
	}

	// Is revoked certificate list is correct?
	for _, serial := range serials {
		found := false
		for _, rc := range rcl {
			//t.Logf("%d == %d", rc.SerialNumber, serial)
			if rc.SerialNumber.Cmp(serial) == 0 {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("revoked serial '%d' is expected to be found in the generated CRL but it is NOT FOUND instead", serial)
		}
	}
}

func TestReadCertFromPEM(t *testing.T) {
	// Initialize:
	ca, _ := pki.NewCA()

	// Prepare:

	// Test:
	crt, err := pki.ReadCertFromPEM(ca.Cert)
	if err != nil {
		t.Fatalf("can not get cert from pem %+v", ca)
	}

	// Is crt nil?
	if crt == nil {
		t.Fatalf("cert is expected to be 'not nil' but it's 'nil' instead")
	}
}

// isPEMEncodedProperly takes an PEM encoded string s and the expected block type typ (e.g. "RSA PRIVATE KEY") and returns whether it can be decodable.
func isPEMEncodedProperly(t *testing.T, s string, typ string) bool {
	block, _ := pem.Decode([]byte(s))

	if block == nil {
		t.Logf("block is nil")
		return false
	}

	if len(block.Bytes) == 0 {
		t.Logf("block bytes length is zero")
		return false
	}

	if block.Type != typ {
		t.Logf("expected block type '%s' but got '%s'", typ, block.Type)
		return false
	}

	switch block.Type {
	case pki.PEMCertificateBlockType:
		crt, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Logf("certificate parse failed %+v: %v", block, err)
			return false
		}

		if crt == nil {
			t.Logf("couldn't parse certificate %+v", block)
			return false
		}
	case pki.PEMRSAPrivateKeyBlockType:
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			t.Logf("private key parse failed %+v: %v", block, err)
			return false
		}

		if key == nil {
			t.Logf("couldn't parse private key %+v", block)
			return false
		}
	case pki.PEMCSRBlockType:
		csr, err := x509.ParseCertificateRequest(block.Bytes)
		if err != nil {
			t.Logf("CSR parse failed %+v: %v", block, err)
			return false
		}

		if csr == nil {
			t.Logf("couldn't parse CSR %+v", block)
			return false
		}

	case pki.PEMx509CRLBlockType:
		crl, err := x509.ParseCRL(block.Bytes)
		if err != nil {
			t.Logf("CRL parse failed %+v: %v", block, err)
			return false
		}

		if crl == nil {
			t.Logf("couldn't parse crl %+v", block)
			return false
		}
	}
	return true
}

// getSerial returns serial number of a pem encoded certificate
func getSerial(t *testing.T, crt string) *big.Int {
	// PEM decode.
	block, _ := pem.Decode([]byte(crt))
	if block == nil {
		t.Fatalf("block is nil %+v", block)
	}

	// Parse certificate.
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("certificate can not be parsed from block %+v: %v", block, err)
	}
	return cert.SerialNumber
}

// randomBetween returns a random int between min and max
func randomBetween(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
