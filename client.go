package main

import (
    "flag"
    "crypto/tls"
    "crypto/x509"
    "os"
    "io"
    "fmt"
    "github.com/op/go-logging"
    "bufio"
)

var (
    debug bool = false
    log *logging.Logger = nil
    port int
    address string
    certFile string
    noverify bool
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
    flag.IntVar(&port, "port", 0, "port to connect")
    flag.StringVar(&address, "address", "", "host to connect to")
    flag.StringVar(&certFile, "certfile", "cert.pem", "trusted CA certificate")
    flag.BoolVar(&noverify, "noverify", false, "do not verify host cert")
    flag.BoolVar(&debug, "debug", false, "debug logging")
    flag.Parse()

    if address == "" || port == 0 {
        log.Error("--address and --port options are both required")
        os.Exit(1)
    }

    log.Infof("connecting to %s:%d", address, port)
    if noverify {
        log.Infof("    TLS verification is disabled")
    }

    cert, err := os.ReadFile(certFile)
    if err != nil {
        panic(err)
    }
    certPool := x509.NewCertPool()
    if ok := certPool.AppendCertsFromPEM(cert); !ok {
        panic("unable to parse cert")
    }
    config := &tls.Config{RootCAs: certPool}
    if noverify {
        config.InsecureSkipVerify = true
    }

    conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", address, port), config)
    if err != nil {
        panic(err)
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
        /*
        if err = conn.CloseWrite(); err != nil {
            log.Errorf("close error: %s", err)
            break
        }
        */

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
