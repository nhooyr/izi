package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"time"
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
	w, err := os.Create(os.Args[2])
	check(err)
	defer w.Close()
	statusc := make(chan string)
	exit := make(chan struct{})
	go printLoop(statusc, exit)
	log.Println("writing to", os.Args[2])
	copyBuffer(w, c, statusc)
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
