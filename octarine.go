package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/dcos/octarine/client"
	"github.com/dcos/octarine/server"
)

var cacheTimeout = flag.Int("cache-timeout", 5,
	"SRV record cache timeout in seconds.")
var verbose = flag.Bool("verbose", false, "Verbose output.")
var cmode = flag.Bool("client", false, "Client mode.")

// Below requires client mode
var queryPort = flag.Bool("port", false,
	"Query the port that's being listened on, only available in client mode.")

func main() {
	flag.Parse()

	sockdir := path.Join(os.TempDir(), "octarine")

	id := flag.Arg(0)
	if id == "" {
		log.Print("Please supply an identifier for this proxy instance")
		os.Exit(1)
	}

	querysock := path.Join(sockdir, fmt.Sprintf("%s.query.sock", id))
	portsock := path.Join(sockdir, fmt.Sprintf("%s.port.sock", id))
	if err := os.MkdirAll(sockdir, os.FileMode(0700)); err != nil {
		log.Fatal(err)
	}

	if *cmode {
		c := &client.Client{
			ID:         id,
			QueryPort:  *queryPort,
			ListenSock: portsock,
			WriteSock:  querysock,
		}
		c.Run()
		os.Exit(0)
	}

	s := &server.Server{
		ID:           id,
		Verbose:      *verbose,
		CacheTimeout: *cacheTimeout,
		ListenSock:   querysock,
		WriteSock:    portsock,
	}
	log.Fatal(s.Run())
}
