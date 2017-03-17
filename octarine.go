package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/dcos/octarine/client"
	"github.com/dcos/octarine/server"
	"github.com/dcos/octarine/util"
)

// XXX Refactor the flags, move them them into the function and use the
//   the parsed version so you get values instead of pointers. Then change
//   remove all the pointer deferences.

var cacheTimeout = flag.Int("cache-timeout", 5,
	"SRV record cache timeout in seconds.")
var verbose = flag.Bool("verbose", false, "Verbose output.")
var cmode = flag.Bool("client", false, "Client mode.")
var proxyMode = flag.String("mode", "",
	fmt.Sprintf("Proxy mode [%s/%s]", server.StandardMode, server.TransparentMode))
var printVersion = flag.Bool("version", false, "Print the version")

// Below requires client mode
var queryPort = flag.Bool("port", false,
	"Query the port that's being listened on, only available in client mode.")

func main() {
	flag.Parse()

	// Short circuit for version
	if *printVersion {
		fmt.Print(strconv.Itoa(util.Version))
		return
	}

	// Validate flags
	id := flag.Arg(0)
	if id == "" {
		log.Fatal("Please supply an identifier for this proxy instance")
	}
	if !*cmode {
		if *proxyMode == "" {
			log.Fatal("Please supply a proxy mode")
		}
		if !server.ValidProxyMode(*proxyMode) {
			log.Fatalf("%s is an invalid proxy mode", *proxyMode)
		}
	}

	sockdir := path.Join(os.TempDir(), "octarine")
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
		ProxyMode:    *proxyMode,
	}
	log.Fatal(s.Run())
}
