package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dcos/mhproxy/srv"
	"github.com/elazarl/goproxy"
)

var dcosDomain = ".mydcos.directory"

var portno = flag.Int("port", 8080, "port to listen on")
var updateInterval = flag.Int("update-interval", 5, "update interval in seconds")
var verbose = flag.Bool("verbose", false, "verbose output")

func dstDomainMatch(domain string) goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		return strings.HasSuffix(req.URL.Host, domain)
	}
}

func stripDcosDomain(r *http.Request, ctx *goproxy.ProxyCtx) (
	*http.Request, *http.Response) {

	r.URL.Host = strings.TrimSuffix(r.URL.Host, dcosDomain)
	return r, nil
}

func createSRVHandler(srvCache srv.SRVCache) func(
	r *http.Request, ctx *goproxy.ProxyCtx) (
	*http.Request, *http.Response) {

	return func(r *http.Request, ctx *goproxy.ProxyCtx) (
		*http.Request, *http.Response) {

		if host, port, err := srvCache.Get(r.URL.Host); err != nil {
			log.Print(err)
		} else {
			r.URL.Host = fmt.Sprintf("%s:%d", host, port)
		}
		return r, nil
	}
}

func main() {
	flag.Parse()
	srvCache := srv.New(time.Duration(*updateInterval) * time.Second)

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest(dstDomainMatch(dcosDomain)).DoFunc(stripDcosDomain)
	srvHandler := createSRVHandler(srvCache)
	proxy.OnRequest().DoFunc(srvHandler)
	if *verbose {
		proxy.Verbose = true
	}
	addr := fmt.Sprintf("127.0.0.1:%d", *portno)
	log.Fatal(http.ListenAndServe(addr, proxy))
}
