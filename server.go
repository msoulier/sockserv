package main

import (
    "flag"
    "crypto/tls"
    "os"
    "net"
    "io"
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
    port := flag.String("port", "4040", "listening port")
    certFile := flag.String("cert", "cert.pem", "certificate PEM file")
    keyFile := flag.String("key", "key.pem", "key PEM file")
    flag.Parse()

    cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
    if err != nil {
        panic(err)
    }
    config := &tls.Config{Certificates: []tls.Certificate{cert}}

    log.Infof("listening on port %s\n", *port)
    l, err := tls.Listen("tcp", ":"+*port, config)
    if err != nil {
        panic(err)
    }
    defer l.Close()

    for {
        conn, err := l.Accept()
        if err != nil {
            panic(err)
        }
        log.Infof("accepted connection from %s\n", conn.RemoteAddr())

        go func(c net.Conn) {
            io.Copy(c, c)
            c.Close()
            log.Errorf("closing connection from %s\n", conn.RemoteAddr())
        }(conn)
    }
}
