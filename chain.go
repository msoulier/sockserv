package main

import (
    "os"
    "fmt"
    "flag"
    "crypto/tls"
    "crypto/x509"
    "encoding/pem"
    "io/ioutil"
    "github.com/op/go-logging"
)

var (
    connect string
    noverify bool
    debug bool
    log *logging.Logger = nil
    path string
)

func init() {
    format := logging.MustStringFormatter(
        `%{time:2006-01-02 15:04:05.000-0700} %{level} [%{shortfile}] %{message}`,  
    )
    stderrBackend := logging.NewLogBackend(os.Stderr, "", 0) 
    stderrFormatter := logging.NewBackendFormatter(stderrBackend, format)
    stderrBackendLevelled := logging.AddModuleLevel(stderrFormatter)
    logging.SetBackend(stderrBackendLevelled)
    if debug {
        stderrBackendLevelled.SetLevel(logging.DEBUG, "chain")
    } else {
        stderrBackendLevelled.SetLevel(logging.INFO, "chain")
    }
    log = logging.MustGetLogger("chain")
}

func loadCerts(path string) ([]*x509.Certificate, error) {
    certs := []*x509.Certificate{}
    certbuf, err := ioutil.ReadFile(path)
    if err != nil {
        return certs, err
    }
    roots := x509.NewCertPool()
    ok := roots.AppendCertsFromPEM([]byte(certbuf))
    if !ok {
        log.Errorf("failed to parse certificate")
        return certs, err
    }
    block, _ := pem.Decode([]byte(certbuf))
	if block == nil {
		log.Error("failed to parse certificate PEM")
        return certs, err
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Errorf("failed to parse certificate: %s", err)
        return certs, err
	}

    certs = append(certs, cert)

    return certs, nil
}

func connectHost(connect string, noverify bool) ([]*x509.Certificate, error) {
    certs := []*x509.Certificate{}
    cfg := tls.Config{}
    if noverify {
        cfg.InsecureSkipVerify = true
    }
    conn, err := tls.Dial("tcp", connect, &cfg)
    if err != nil {
        log.Fatal("TLS connection failed: " + err.Error())
        return certs, err
    }
    defer conn.Close()

    certs = conn.ConnectionState().PeerCertificates
    return certs, nil
}

func main() {
    flag.StringVar(&connect, "connect", "", "address:port to connect to")
    flag.StringVar(&path, "path", "", "path on disk to a PEM file to check")
    flag.BoolVar(&noverify, "noverify", false, "do not verify host cert")
    flag.Parse()

    if connect == "" && path == "" {
        log.Error("either --connect or --path are required options")
        os.Exit(1)
    }

    certChain := []*x509.Certificate{}

    var err error = nil
    if connect != "" {
        certChain, err = connectHost(connect, noverify)
        if err != nil {
            log.Errorf("connectHost: %s", err)
            os.Exit(1)
        }
    } else if path != "" {
        certChain, err = loadCerts(path)
        if err != nil {
            log.Errorf("loadCerts: %s", err)
            os.Exit(1)
        }
    } else {
        panic("we should not be here")
    }

    for i, cert := range certChain {
        fmt.Println(i)
        fmt.Println("Issuer:", cert.Issuer)
        fmt.Println("Subject:", cert.Subject)
        fmt.Println("Version:", cert.Version)
        fmt.Println("NotAfter:", cert.NotAfter)
        fmt.Println("DNS names:", cert.DNSNames)
        fmt.Println("")
    }
}
