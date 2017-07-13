package ovpm

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	//"github.com/Sirupsen/logrus"
	"bytes"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// CA represents x509 Certificate Authority.
type CA struct {
	Cert string
	Key  string // Private Key
	CSR  string
}

// Cert represents any certificate - key pair.
type Cert struct {
	Cert string
	Key  string // Private Key
}

// CreateCA generates a certificate and a key-pair for the CA and returns them.
func CreateCA() (*CA, error) {
	key, err := rsa.GenerateKey(rand.Reader, CrtKeyLength)
	if err != nil {
		return nil, fmt.Errorf("private key cannot be created: %s", err)
	}

	val, err := asn1.Marshal(basicConstraints{true, 0})
	if err != nil {
		return nil, fmt.Errorf("can not marshal basic constraints: %s", err)
	}

	names := pkix.Name{CommonName: "CA"}
	var csrTemplate = x509.CertificateRequest{
		Subject:            names,
		SignatureAlgorithm: x509.SHA512WithRSA,
		ExtraExtensions: []pkix.Extension{
			{
				Id:       asn1.ObjectIdentifier{2, 5, 29, 19},
				Value:    val,
				Critical: true,
			},
		},
	}

	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, key)
	if err != nil {
		return nil, fmt.Errorf("can not create certificate request: %s", err)
	}

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	// Serial number
	serial, err := rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(159), nil))
	if err != nil {
		return nil, err
	}

	now := time.Now()
	// Create the request template
	template := x509.Certificate{
		SerialNumber:          serial,
		Subject:               names,
		NotBefore:             now.Add(-10 * time.Minute).UTC(),
		NotAfter:              now.Add(time.Duration(24*365) * time.Hour).UTC(),
		BasicConstraintsValid: true,
		IsCA:     true,
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		//ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	// Sign the certificate authority
	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate error: %s", err)
	}

	var request bytes.Buffer
	var privateKey bytes.Buffer
	if err := pem.Encode(&request, &pem.Block{Type: "CERTIFICATE", Bytes: certificate}); err != nil {
		return nil, err
	}
	if err := pem.Encode(&privateKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		return nil, err
	}

	return &CA{
		Key:  privateKey.String(),
		Cert: request.String(),
		CSR:  string(csr),
	}, nil

}

// CreateServerCert generates a x509 certificate and a key-pair for the server.
func CreateServerCert(ca *CA) (*Cert, error) {
	return createCert("localhost", ca, true)
}

// CreateClientCert generates a x509 certificate and a key-pair for the client.
func CreateClientCert(username string, ca *CA) (*Cert, error) {
	return createCert(username, ca, false)
}

func getCertFromPEM(pemCert string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(pemCert))
	var cert *x509.Certificate
	cert, _ = x509.ParseCertificate(block.Bytes)
	return cert, nil
}

func createCert(commonName string, ca *CA, server bool) (*Cert, error) {
	// Get CA private key
	block, _ := pem.Decode([]byte(ca.Key))
	if block == nil {
		return nil, fmt.Errorf("failed to parse ca private key")
	}

	caKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ca private key: %s", err)
	}

	caCert, err := getCertFromPEM(ca.Cert)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ca cert: %v", err)
	}

	// Create new cert's key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("private key cannot be created: %s", err)
	}

	serial, err := rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(159), nil))
	if err != nil {
		return nil, err
	}

	tml := x509.Certificate{
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(5, 0, 0),
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Innovation"},
		},
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	if server {
		tml.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}

	// Sign with CA's private key
	cert, err := x509.CreateCertificate(rand.Reader, &tml, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("certificate cannot be created: %s", err)
	}

	priKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	return &Cert{
		Key:  string(priKeyPem[:]),
		Cert: string(certPem[:]),
	}, nil
}

type basicConstraints struct {
	IsCA       bool `asn1:"optional"`
	MaxPathLen int  `asn1:"optional,default:-1"`
}

func getCA() (*CA, error) {
	server := Server{}
	db.First(&server)
	if db.NewRecord(&server) {
		return nil, fmt.Errorf("can not retrieve server from db")
	}
	return &CA{
		Cert: server.CACert,
		Key:  server.CAKey,
	}, nil

}
