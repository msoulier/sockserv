package main

import (
    "fmt"
    "flag"
    "crypto/tls"
    "log"
)

func main() {
    addr := flag.String("addr", "localhost:4040", "dial address")
    noverify := flag.Bool("noverify", false, "do not verify host cert")
    flag.Parse()

    cfg := tls.Config{}
    if *noverify {
        cfg.InsecureSkipVerify = true
    }
    conn, err := tls.Dial("tcp", *addr, &cfg)
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
