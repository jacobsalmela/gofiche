package main

import (
	"flag"
	"fmt"

	"github.com/jacobsalmela/gofiche"
)

var (
	listenAddr = "0.0.0.0"
	port       = 9999
	domain     = "localhost"
	slugLen    = 6
	outDir     = "pastes"
	debug      = false
	bufSize    int
	userName   string
	logFile    string
)

func main() {
	flag.IntVar(&slugLen, "slug", slugLen, "Length of the randomly-generated slug")
	flag.IntVar(&port, "port", port, "Port the server will listen on")
	flag.StringVar(&listenAddr, "listen", listenAddr, fmt.Sprintf("Address on which gofiche is waiting for connections"))
	flag.StringVar(&domain, "domain", domain, "Domain used in output lines")
	flag.BoolVar(&debug, "debug", debug, "Show additional server side logging")
	flag.Parse()
	slug := gofiche.Slug{
		Length: slugLen,
	}

	settings := &gofiche.GoficheSettings{
		ListenAddr: listenAddr,
		Port:       port,
		Slug:       slug,
		OutDir:     outDir,
		Debug:      debug,
	}
	settings.Domain = fmt.Sprintf("%s", domain)
	gofiche.Serve(settings)
}
