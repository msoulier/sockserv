package main

import (
    "os"
    "fmt"
    "flag"
    "crypto/tls"
    "github.com/op/go-logging"
)

var (
    connect string
    noverify bool
    debug bool
    log *logging.Logger = nil
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

func main() {
    flag.StringVar(&connect, "connect", "", "address:port to connect to")
    flag.BoolVar(&noverify, "noverify", false, "do not verify host cert")
    flag.Parse()

    if connect == "" {
        log.Error("--connect is a required option")
        os.Exit(1)
    }

    cfg := tls.Config{}
    if noverify {
        cfg.InsecureSkipVerify = true
    }
    conn, err := tls.Dial("tcp", connect, &cfg)
    if err != nil {
        log.Fatal("TLS connection failed: " + err.Error())
    }
    defer conn.Close()

    certChain := conn.ConnectionState().PeerCertificates
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
