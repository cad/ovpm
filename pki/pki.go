// Package pki contains bits and pieces to work with OpenVPN PKI related operations.
package pki

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

const (
	_CrtExpireYears = 10
	_CrtKeyLength   = 2024
)

// CertHolder encapsulates a public certificate and the corresponding private key.
type CertHolder struct {
	Cert string // PEM Encoded Certificate
	Key  string // PEM Encoded Private Key
}

// CA is a special type of CertHolder that also has a CSR in it.
type CA struct {
	CertHolder
	CSR string
}

// NewCA returns a newly generated CA.
//
// This will generate a public/private RSA keypair and a authority certificate signed by itself.
func NewCA() (*CA, error) {
	type basicConstraints struct {
		IsCA       bool `asn1:"optional"`
		MaxPathLen int  `asn1:"optional,default:-1"`
	}

	key, err := rsa.GenerateKey(rand.Reader, _CrtKeyLength)
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
		Type: PEMCSRBlockType, Bytes: csrCertificate,
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
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		//ExtKeyUsage: []x509.ExtKeyUsage{x509.KeyUsageCertSign, x509.ExtKeyUsageClientAuth},
	}

	// Sign the certificate authority
	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate error: %s", err)
	}

	var request bytes.Buffer
	var privateKey bytes.Buffer
	if err := pem.Encode(&request, &pem.Block{Type: PEMCertificateBlockType, Bytes: certificate}); err != nil {
		return nil, err
	}
	if err := pem.Encode(&privateKey, &pem.Block{Type: PEMRSAPrivateKeyBlockType, Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		return nil, err
	}

	return &CA{
		CertHolder: CertHolder{
			Key:  privateKey.String(),
			Cert: request.String(),
		},
		CSR: string(csr),
	}, nil

}

// NewServerCertHolder generates a RSA key-pair and a x509 certificate signed by the CA for the server.
func NewServerCertHolder(ca *CA) (*CertHolder, error) {
	return newCert(ca, true, "localhost")
}

// NewClientCertHolder generates a RSA key-pair and a x509 certificate signed by the CA for the client.
func NewClientCertHolder(ca *CA, username string) (*CertHolder, error) {
	return newCert(ca, false, username)
}

// newCert generates a RSA key-pair and a x509 certificate signed by the CA.
func newCert(ca *CA, server bool, cn string) (*CertHolder, error) {
	// Get CA private key
	block, _ := pem.Decode([]byte(ca.Key))
	if block == nil {
		return nil, fmt.Errorf("failed to parse ca private key")
	}

	caKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ca private key: %s", err)
	}

	caCert, err := ReadCertFromPEM(ca.Cert)
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
			CommonName:   cn,
			Organization: []string{"OVPM"},
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
		Type:  PEMRSAPrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  PEMCertificateBlockType,
		Bytes: cert,
	})

	return &CertHolder{
		Key:  string(priKeyPem[:]),
		Cert: string(certPem[:]),
	}, nil
}

// NewCRL takes in a list of certificate serial numbers to-be-revoked and a CA then makes a PEM encoded CRL and returns it as a string.
func NewCRL(ca *CA, serials ...*big.Int) (string, error) {
	caCrt, err := ReadCertFromPEM(ca.Cert)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode([]byte(ca.Key))
	if block == nil {
		return "", fmt.Errorf("failed to parse ca private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse ca private key: %s", err)
	}
	var revokedCertList []pkix.RevokedCertificate
	for _, serial := range serials {
		revokedCert := pkix.RevokedCertificate{
			SerialNumber:   serial,
			RevocationTime: time.Now().UTC(),
		}
		revokedCertList = append(revokedCertList, revokedCert)
	}
	crl, err := caCrt.CreateCRL(rand.Reader, priv, revokedCertList, time.Now().UTC(), time.Now().Add(365*24*60*time.Minute).UTC())
	if err != nil {
		return "", err
	}

	crlPem := pem.EncodeToMemory(&pem.Block{
		Type:  PEMx509CRLBlockType,
		Bytes: crl,
	})

	return string(crlPem[:]), nil

}

// ReadCertFromPEM decodes a PEM encoded string into a x509.Certificate.
func ReadCertFromPEM(s string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(s))
	var cert *x509.Certificate
	cert, _ = x509.ParseCertificate(block.Bytes)
	return cert, nil
}
