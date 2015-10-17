package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
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
	logc := make(chan string)
	exit := make(chan struct{})
	go logLoop(logc, exit)
	log.Println("sending", os.Args[2])
	copyBuffer(c, r, logc)
	<-exit
	log.Println("done")
}

func logLoop(logc chan string, exit chan struct{}) {
	for m := range logc {
		log.Println(m)
	}
	exit <- struct{}{}
}

func copyBuffer(dst io.Writer, src io.Reader, logc chan string) {
	defer close(logc)
	buf := make([]byte, 50000000)
	for {
		nr, er := src.Read(buf)
		nw, ew := dst.Write(buf[0:nr])
		if nw > 0 {
			logc <- fmt.Sprintf("sent %d bytes", nw)
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
