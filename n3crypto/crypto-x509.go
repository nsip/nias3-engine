// crypto-x509.go

package n3crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

//
// creates a new self-signed X.509 certificate if not already there
//

// https://gist.github.com/jaymecd/2445f516d7d2635eb394

func publicKey(priv *ecdsa.PrivateKey) interface{} {
	return &priv.PublicKey
}

func pemBlockForKey(priv *ecdsa.PrivateKey) *pem.Block {
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
		os.Exit(2)
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
}

var hostnames = "localhost"
var organizationName = "Localhost"
var organizationUnit = "Localhost"
var expiryYears = 10

func NewTLSCertificate() {

	_, errCert := ioutil.ReadFile("cert.pem")
	_, errKey := ioutil.ReadFile("key.pem")

	if errKey == nil && errCert == nil {
		return
	}

	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{organizationName},
			OrganizationalUnit: []string{organizationUnit},
			CommonName:         fmt.Sprintf("INTERNAL TESTING (%s)", organizationName),
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(expiryYears, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:        true,
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	hosts := strings.Split(hostnames, ",")
	for _, h := range hosts {
		template.DNSNames = append(template.DNSNames, h)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}
	/*
		rootCert, err := x509.ParseCertificate(derBytes)
		if err != nil {
			return
		}
	*/

	// https://ericchiang.github.io/post/go-tls/

	certOut, err := os.Create("cert.pem")
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	log.Print("written cert.pem\n")

	keyOut, err := os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Print("failed to open key.pem for writing:", err)
		return
	}
	pem.Encode(keyOut, pemBlockForKey(priv))
	keyOut.Close()
	log.Print("written key.pem\n")
}
