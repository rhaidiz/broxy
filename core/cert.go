// thanks to evilsocket's amazing code from bettercap

package core

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
)

var (
	certCache = make(map[string]*tls.Certificate)
	certLock  = &sync.Mutex{}
)

// CreateCA returns a Certificate Authority and its private key
func CreateCA() ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	notBefore := time.Now()
	aYear := time.Duration(365*24) * time.Hour
	notAfter := notBefore.Add(aYear)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"US"},
			Locality:           []string{""},
			Organization:       []string{"broxy"},
			OrganizationalUnit: []string{"https://github.com/rhaidiz/broxy/"},
			CommonName:         "broxy Certificate Authority",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	return x509.MarshalPKCS1PrivateKey(priv), cert, err
}

// TLSConfigFromCA returns a TLS configuration with a custom CA
func TLSConfigFromCA(ca *tls.Certificate) func(host string, ctx *goproxy.ProxyCtx) (*tls.Config, error) {
	return func(host string, ctx *goproxy.ProxyCtx) (c *tls.Config, err error) {
		parts := strings.SplitN(host, ":", 2)
		hostname := parts[0]
		port := 443
		if len(parts) > 1 {
			port, err = strconv.Atoi(parts[1])
			if err != nil {
				port = 443
			}
		}

		cert := getCachedCert(hostname, port)
		if cert == nil {
			cert, err = signHost(ca, hostname, port)
			if err != nil {
				return nil, err
			}
			setCachedCert(hostname, port, cert)
		}

		config := tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{*cert},
		}

		return &config, nil
	}
}

func signHost(ca *tls.Certificate, host string, port int) (cert *tls.Certificate, err error) {
	var x509ca *x509.Certificate
	var template x509.Certificate

	if x509ca, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return
	}

	notBefore := time.Now()
	aYear := time.Duration(365) * time.Hour
	notAfter := notBefore.Add(aYear)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template = x509.Certificate{
		SerialNumber: serialNumber,
		Issuer:       x509ca.Subject,
		Subject: pkix.Name{
			Country:            []string{"US"},
			Locality:           []string{""},
			Organization:       []string{"broxy"},
			OrganizationalUnit: []string{"https://github.com/rhaidiz/broxy/"},
			CommonName:         "broxy mitm certificate",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		//KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		//ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		//BasicConstraintsValid: true,
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}

	var certpriv *rsa.PrivateKey
	if certpriv, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return
	}

	var derBytes []byte
	if derBytes, err = x509.CreateCertificate(rand.Reader, &template, x509ca, &certpriv.PublicKey, ca.PrivateKey); err != nil {
		return
	}

	return &tls.Certificate{
		Certificate: [][]byte{derBytes, ca.Certificate[0]},
		PrivateKey:  certpriv,
	}, nil
}

func keyFor(domain string, port int) string {
	return fmt.Sprintf("%s:%d", domain, port)
}

func getCachedCert(domain string, port int) *tls.Certificate {
	certLock.Lock()
	defer certLock.Unlock()
	if cert, found := certCache[keyFor(domain, port)]; found {
		return cert
	}
	return nil
}

func setCachedCert(domain string, port int, cert *tls.Certificate) {
	certLock.Lock()
	defer certLock.Unlock()
	certCache[keyFor(domain, port)] = cert
}
