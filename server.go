package main

import (
    "flag"
    "crypto/tls"
    "os"
    "net"
    "github.com/op/go-logging"
    "bufio"
    "fmt"
)

var (
    debug bool = false
    log *logging.Logger = nil
    port int
    certFile string
    keyFile string
    listen string
    noecho bool
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
    flag.StringVar(&listen, "listen", "0.0.0.0", "IP to listen on")
    flag.IntVar(&port, "port", 0, "listening port")
    flag.StringVar(&certFile, "cert", "cert.pem", "certificate PEM file")
    flag.StringVar(&keyFile, "key", "key.pem", "key PEM file")
    flag.BoolVar(&debug, "debug", false, "debug logging")
    flag.BoolVar(&noecho, "noecho", false, "do not echo input back on the socket")
    flag.Parse()

    if port == 0 {
        log.Error("--port is a required option")
        os.Exit(1)
    }

    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        panic(err)
    }
    config := &tls.Config{Certificates: []tls.Certificate{cert}}

    log.Infof("listening on %s:%d", listen, port)
    l, err := tls.Listen("tcp", fmt.Sprintf("%s:%d", listen, port), config)
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
            defer c.Close()
            for {
                buffer, err := bufio.NewReader(c).ReadString('\n')
                if err != nil {
                    log.Errorf("read error: %s", err)
                    break
                }
                log.Infof("received: %s", buffer)
                if !noecho {
                    fmt.Fprintf(c, buffer)
                }
            }
            log.Errorf("closing connection from %s\n", conn.RemoteAddr())
        }(conn)
    }
}
