package main

import (
    "flag"
    "crypto/tls"
    "crypto/x509"
    "os"
    "io"
    "fmt"
    "github.com/op/go-logging"
)

var (
    debug bool = false
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
        stderrBackendLevelled.SetLevel(logging.DEBUG, "sockserv")
    } else {
        stderrBackendLevelled.SetLevel(logging.INFO, "sockserv")
    }
    log = logging.MustGetLogger("sockserv")
}

func main() {
    port := flag.String("port", "4040", "port to connect")
    certFile := flag.String("certfile", "cert.pem", "trusted CA certificate")
    noverify := flag.Bool("noverify", false, "do not verify host cert")
    flag.Parse()

    cert, err := os.ReadFile(*certFile)
    if err != nil {
        panic(err)
    }
    certPool := x509.NewCertPool()
    if ok := certPool.AppendCertsFromPEM(cert); !ok {
        panic("unable to parse cert")
    }
    config := &tls.Config{RootCAs: certPool}
    if *noverify {
        config.InsecureSkipVerify = true
    }

    conn, err := tls.Dial("tcp", "localhost:"+*port, config)
    if err != nil {
        panic(err)
    }

    _, err = io.WriteString(conn, "Hello simple secure Server\n")
    if err != nil {
        panic(err)
    }
    if err = conn.CloseWrite(); err != nil {
        panic(err)
    }

    buf := make([]byte, 256)
    n, err := conn.Read(buf)
    if err != nil && err != io.EOF {
        panic(err)
    }

    fmt.Println("client read:", string(buf[:n]))
    conn.Close()
}
