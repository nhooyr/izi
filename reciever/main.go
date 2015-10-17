package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	log.SetPrefix("reciever: ")
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
	config.InsecureSkipVerify = true
	config.CipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	}
	log.Println("connecting")
	c, err := tls.Dial("tcp", os.Args[1], config)
	check(err)
	defer c.Close()
	w, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
	check(err)
	defer w.Close()
	logc := make(chan string)
	exit := make(chan struct{})
	go logLoop(logc, exit)
	log.Println("writing to", os.Args[2])
	copyBuffer(w, c, logc)
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
		logc <- fmt.Sprintf("got %d bytes", nw)
		}
		if er == io.EOF {
			return
		} else if er != nil {
			panic(er)
		}
		if ew != nil {
			panic(ew)
		}
	}
}
