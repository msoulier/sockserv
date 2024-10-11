package main

import (
    "bufio"
    "crypto/tls"
    "crypto/x509"
    "errors"
    "flag"
    "fmt"
    "io"
    "net"
    "os"

    "github.com/op/go-logging"
)

var (
    debug bool = false
    log *logging.Logger = nil
    connect string
    certFile string
    noverify bool
    transport string
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

func getConnTCP(connect string) (net.Conn, error) {
    conn, err := net.Dial("tcp", connect)
    if err != nil {
        log.Error("Dial: %s", err)
        return conn, err
    }
    return conn, nil
}

func getConnTLS(connect, certFile string, noverify bool) (net.Conn, error) {
    var conn net.Conn
    cert, err := os.ReadFile(certFile)
    if err != nil {
        log.Errorf("ReadFile: %s", err)
        return conn, err
    }
    certPool := x509.NewCertPool()
    if ok := certPool.AppendCertsFromPEM(cert); !ok {
        log.Error("unable to parse cert")
        return conn, errors.New("unable to parse cert")
    }
    config := &tls.Config{RootCAs: certPool}
    if noverify {
        config.InsecureSkipVerify = true
    }

    conn, err = tls.Dial("tcp", connect, config)
    if err != nil {
        log.Errorf("Dial: %s", err)
        return conn, err
    }

    return conn, nil
}

func main() {
    flag.StringVar(&connect, "connect", "", "host:port to connect to")
    flag.StringVar(&certFile, "certfile", "cert.pem", "trusted CA certificate")
    flag.StringVar(&transport, "transport", "tcp", "transport to use (tcp|tls)")
    flag.BoolVar(&noverify, "noverify", false, "do not verify host cert")
    flag.BoolVar(&debug, "debug", false, "debug logging")
    flag.Parse()

    if connect == "" {
        flag.PrintDefaults()
        log.Error("--connect is a required option")
        os.Exit(1)
    }

    log.Infof("connecting to %s", connect)
    if noverify {
        log.Infof("    TLS verification is disabled")
    }

    var conn net.Conn
    var err error
    if transport == "tcp" {
        conn, err = getConnTCP(connect)
        if err != nil {
            log.Errorf("getConnTCP: %s", err)
            os.Exit(1)
        }
    } else if transport == "tls" {
        conn, err = getConnTLS(connect, certFile, noverify)
        if err != nil {
            log.Errorf("getConnTLS: %s", err)
            os.Exit(1)
        }
    }

    stdinReader := bufio.NewReader(os.Stdin)

    for {
        fmt.Print("> ")
        if inbuf, err := stdinReader.ReadString('\n'); err != nil {
            log.Errorf("error reading from stdin: %s", err)
            break
        } else {
            n, err := io.WriteString(conn, inbuf)
            if err != nil {
                log.Errorf("write error: %s", err)
                break
            }
            log.Infof("wrote %d bytes to server", n)
        }

        buf := make([]byte, 256)
        n, err := conn.Read(buf)
        if err != nil && err != io.EOF {
            log.Errorf("read error: %s", err)
            break
        } else if err != nil && err == io.EOF {
            log.Errorf("EOF")
            break
        }

        fmt.Println("client read:", string(buf[:n]))
    }
    conn.Close()
}
