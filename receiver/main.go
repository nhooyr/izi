package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	GIGABYTE = 1000000000
	MEGABYTE = 1000000
	KILOBYTE = 1000
	CSI      = "\033["
	DEL1     = CSI + "1M"
	REDRAW   = DEL1 + "\r"
	PREFIX   = "receiver:"
)

func println(v ...interface{}) {
	fmt.Println(append([]interface{}{PREFIX}, v...)...)
}

func main() {
	log.SetPrefix(PREFIX + " ")
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
	println("connecting")
	c, err := tls.Dial("tcp", os.Args[1], config)
	check(err)
	defer c.Close()
	w, err := os.Create(os.Args[2])
	check(err)
	defer w.Close()
	statusc := make(chan float64)
	exit := make(chan struct{})
	go statusLoop(statusc, exit)
	println("writing to", os.Args[2])
	copyTo(c, w, statusc)
	<-exit
	fmt.Println("\n"+PREFIX, "done")
}

func statusLoop(statusc chan float64, exit chan struct{}) {
	for avg := range statusc {
		var u string
		switch {
		case avg >= GIGABYTE:
			u = fmt.Sprintf("%f gigabytes per second", avg/GIGABYTE)
		case avg >= MEGABYTE:
			u = fmt.Sprintf("%f megabytes per second", avg/MEGABYTE)
		case avg >= KILOBYTE:
			u = fmt.Sprintf("%f kilobytes per second", avg/KILOBYTE)
		default:
			u = fmt.Sprintf("%f bytes per second", avg)
		}
		fmt.Print(REDRAW+PREFIX, " "+u)
	}
	exit <- struct{}{}
}

func copyTo(src io.Reader, dst io.Writer, statusc chan float64) {
	defer close(statusc)
	buf := make([]byte, 32*1024)
	var ns int
	var last float64
	start := time.Now()
	for {
		nr, er := src.Read(buf)
		nw, ew := dst.Write(buf[:nr])
		if nw > 0 {
			ns += nw
		}
		if ew != nil {
			panic(ew)
		}
		if er == io.EOF {
			statusc <- float64(ns) / time.Since(start).Seconds()
			return
		} else if er != nil {
			panic(er)
		}
		if now := time.Since(start).Seconds(); now > last+1 {
			last = now
			select {
			case statusc <- float64(ns) / now:
			default:
			}
		}
	}
}
