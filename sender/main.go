package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	log.SetPrefix("sender: ")
	log.SetFlags(0)
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	if len(os.Args) < 3 {
		log.Fatal("give arguments plz")
	}
	config := new(tls.Config)
	config.CipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	}
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	check(err)
	if cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0]); err != nil {
		log.Fatal(err)
	}
	config.Certificates = []tls.Certificate{cert}
	l, err := tls.Listen("tcp", os.Args[1], config)
	check(err)
	defer l.Close()
	log.Println("listening...")
	c, err := l.Accept()
	check(err)
	defer c.Close()
	log.Println("got connection from", c.RemoteAddr().String())
	r, err := os.OpenFile(os.Args[2], os.O_RDONLY|os.O_SYNC, 0)
	check(err)
	defer r.Close()
	statusc := make(chan string)
	exit := make(chan struct{})
	go printLoop(statusc, exit)
	log.Println("sending", os.Args[2])
	copyBuffer(c, r, statusc)
	<-exit
	log.Println("done")
}

func printLoop(statusc chan string, exit chan struct{}) {
	for m := range statusc {
		fmt.Printf(REDRAW + m)
	}
	exit <- struct{}{}
}

const (
	GIGABYTE = 1000000000
	MEGABYTE = 1000000
	KILOBYTE = 1000
	CSI      = "\033["
	REDRAW   = CSI + "1M\r"
)

func copyBuffer(dst io.Writer, src io.Reader, statusc chan string) {
	defer close(statusc)
	buf := make([]byte, 16*1024)
	start := time.Now()
	var ns int
	var last float64
	for {
		nr, er := src.Read(buf)
		nw, ew := dst.Write(buf[0:nr])
		if nw > 0 {
			ns += nw
			if now := time.Since(start).Seconds(); now > last+1 {
				last = now
				avg := float64(ns) / now
				statusc <- REDRAW
				switch {
				case avg >= GIGABYTE:
					statusc <- fmt.Sprintf("%f gigabytes per second", avg/GIGABYTE)
				case avg >= MEGABYTE:
					statusc <- fmt.Sprintf("%f megabytes per second", avg/MEGABYTE)
				case avg >= KILOBYTE:
					statusc <- fmt.Sprintf("%f kilobytes per second", avg/KILOBYTE)
				default:
					statusc <- fmt.Sprintf("%f bytes per second", avg)
				}
			}
		}
		if ew != nil {
			panic(ew)
		}
		if er == io.EOF {
			return
		} else if er != nil {
			panic(er)
		}
	}
}
