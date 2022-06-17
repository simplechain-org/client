package safe

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

const (
	defaultClientAuth = "requireandverifyclientcert"
)

var clientAuthTypes = map[string]tls.ClientAuthType{
	"noclientcert":               tls.NoClientCert,
	"requestclientcert":          tls.RequestClientCert,
	"requireanyclientcert":       tls.RequireAnyClientCert,
	"verifyclientcertifgiven":    tls.VerifyClientCertIfGiven,
	"requireandverifyclientcert": tls.RequireAndVerifyClientCert,
}
var defaultCipherSuites = []uint16{
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
}

func NewTLSServerListener(endpoint string, certFile string, keyFile string, Type string, certFiles []string) (net.Listener, error) {
	var clientAuth tls.ClientAuthType
	var ok bool
	var listener net.Listener
	// If key file is specified and it does not exist or its corresponding certificate file does not exist
	// then need to return error and not start the server. The stls key file is specified when the user
	// wants the server to use custom stls key and cert and don't want server to auto generate its own. So,
	// when the key file is specified, it must exist on the file system
	if keyFile != "" {
		if !FileExists(keyFile) {
			return nil, fmt.Errorf("file specified by 'tls.keyfile' does not exist: %s", keyFile)
		}
		if !FileExists(certFile) {
			return nil, fmt.Errorf("file specified by 'tls.certfile' does not exist: %s", certFile)
		}
		fmt.Println("tls Certificate:", certFile, "  stls Key: ", keyFile)
	} else if !FileExists(certFile) {
		// stls key file is not specified, generate stls key and cert if they are not already generated
		return nil, fmt.Errorf("failed to tls certificate and key file are not existed")
	}

	cer, err := LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	if Type == "" {
		Type = defaultClientAuth
	}

	fmt.Println("Client authentication type requested:", Type)

	authType := strings.ToLower(Type)
	if clientAuth, ok = clientAuthTypes[authType]; !ok {
		return nil, fmt.Errorf("invalid client auth type provided")
	}

	var certPool *x509.CertPool
	if authType == defaultClientAuth {
		certPool, err = LoadPEMCertPool(certFiles)
		if err != nil {
			return nil, err
		}
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{*cer},
		ClientAuth:   clientAuth,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
		CipherSuites: defaultCipherSuites,
	}

	listener, err = tls.Listen("tcp", endpoint, config)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func NewTLSClient(certFile string, keyFile string, certFiles []string) (*http.Client, error) {
	tr := new(http.Transport)
	var certs []tls.Certificate
	fmt.Println("CA Files:", certFiles)
	fmt.Println("Client Cert File:", certFile)
	fmt.Println("Client Key File:", keyFile)
	if certFile != "" {
		var err = CheckCertDates(certFile)
		if err != nil {
			return nil, err
		}
		clientCert, err := LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		certs = append(certs, *clientCert)
	} else {
		fmt.Println("Client TLS certificate and/or key file not provided")
	}
	rootCAPool := x509.NewCertPool()
	if len(certFiles) == 0 {
		return nil, fmt.Errorf("no trusted root certificates for TLS were provided")
	}

	for _, cacert := range certFiles {
		caCert, err := ioutil.ReadFile(cacert)
		if err != nil {
			return nil, fmt.Errorf("failed to read '%s %v'", cacert, err)
		}
		ok := rootCAPool.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, fmt.Errorf("failed to process certificate from file %s", cacert)
		}
	}
	tlsConfig := &tls.Config{
		Certificates: certs,
		RootCAs:      rootCAPool,
	}
	// set the default ciphers
	tlsConfig.CipherSuites = defaultCipherSuites
	tr.TLSClientConfig = tlsConfig
	tr.TLSClientConfig.InsecureSkipVerify = true
	httpClient := &http.Client{Transport: tr}
	return httpClient, nil
}

func NewTLSClientConfig(certFile string, keyFile string, certFiles []string) (*tls.Config, error) {
	var certs []tls.Certificate
	fmt.Println("CA Files:", certFiles)
	fmt.Println("Client Cert File:", certFile)
	fmt.Println("Client Key File:", keyFile)
	if certFile != "" {
		var err = CheckCertDates(certFile)
		if err != nil {
			return nil, err
		}
		clientCert, err := LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		certs = append(certs, *clientCert)
	} else {
		fmt.Println("Client TLS certificate and/or key file not provided")
	}
	rootCAPool := x509.NewCertPool()
	if len(certFiles) == 0 {
		return nil, fmt.Errorf("no trusted root certificates for TLS were provided")
	}

	for _, cacert := range certFiles {
		caCert, err := ioutil.ReadFile(cacert)
		if err != nil {
			return nil, fmt.Errorf("failed to read '%s %v'", cacert, err)
		}
		ok := rootCAPool.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, fmt.Errorf("failed to process certificate from file %s", cacert)
		}
	}
	tlsConfig := &tls.Config{
		Certificates:       certs,
		RootCAs:            rootCAPool,
		InsecureSkipVerify: true,
	}
	// set the default ciphers
	tlsConfig.CipherSuites = defaultCipherSuites
	return tlsConfig, nil
}
