package certmanager

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	log "github.com/sirupsen/logrus"
	"math/big"
	"time"
)

// certManager helps to generate certs
type certManager struct {
	Organizations []string      `json:"organizations"`
	EffectiveTime time.Duration `json:"effectiveTime"`
	DNSNames      []string      `json:"DNSNames"`
	CommonName    string        `json:"commonName"`
}

func NewCertManager(
	Orz []string,
	effectiveTime time.Duration,
	dnsNames []string,
	commonName string) *certManager {
	return &certManager{
		Organizations: Orz,
		EffectiveTime: effectiveTime,
		DNSNames:      dnsNames,
		CommonName:    commonName,
	}
}

// GenerateSelfSignedCerts return self-signed certs according to provided dns
func (m *certManager) GenerateSelfSignedCerts() (serverCertPEM *bytes.Buffer, serverPrivateKeyPEM *bytes.Buffer, err error) {
	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization: m.Organizations,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(m.EffectiveTime),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	var caPrivateKey *rsa.PrivateKey
	caPrivateKey, err = rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		log.WithError(err).Error("failed to generate key")
		return
	}

	// self signed CA certificate
	var caBytes []byte
	caBytes, err = x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.WithError(err).Error("failed to create certs")
		return
	}

	// PEM encode CA cert
	var caPEM = new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     m.DNSNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   m.CommonName,
			Organization: m.Organizations,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	var serverPrivateKey *rsa.PrivateKey
	serverPrivateKey, err = rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		log.WithError(err).Error("failed to generate server private key")
		return
	}

	// sign the server cert
	var serverCertBytes []byte
	serverCertBytes, err = x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.WithError(err).Error("failed to generate server public cert")
		return
	}

	// PEM encode the server cert and key
	serverCertPEM = new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	serverPrivateKeyPEM = new(bytes.Buffer)
	_ = pem.Encode(serverPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})

	return
}
