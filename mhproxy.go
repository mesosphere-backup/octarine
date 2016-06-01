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

const dcosDomain string = ".mydcos.directory"

var portno = flag.Int("port", 8080, "port to listen on")
var cacheTimeout = flag.Int("cache-timeout", 5, "SRV record cache timeout in seconds")
var verbose = flag.Bool("verbose", false, "verbose output")

func dstSuffixMatch(suffix string) goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		return strings.HasSuffix(req.URL.Host, suffix)
	}
}

func dstFirstCharMatch(char byte) goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		return req.URL.Host[0] == char
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

func createNonProxyHandler(proxy *goproxy.ProxyHttpServer,
	trafficType string) func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {
		if req.Host == "" {
			msg := "Cannot handle requests without Host header, e.g., HTTP 1.0"
			fmt.Fprintln(w, msg)
			return
		}
		req.URL.Scheme = trafficType
		req.URL.Host = req.Host
		proxy.ServeHTTP(w, req)
	}
}

func main() {
	flag.Parse()
	srvCache := srv.New(time.Duration(*cacheTimeout) * time.Second)
	srvHandler := createSRVHandler(srvCache)

	proxy := goproxy.NewProxyHttpServer()
	httpProxifier := createNonProxyHandler(proxy, "http")
	proxy.NonproxyHandler = http.HandlerFunc(httpProxifier)
	proxy.OnRequest(dstSuffixMatch(dcosDomain)).DoFunc(stripDcosDomain)
	proxy.OnRequest(dstFirstCharMatch("_"[0])).DoFunc(srvHandler)
	proxy.Verbose = *verbose

	addr := fmt.Sprintf("127.0.0.1:%d", *portno)
	log.Fatal(http.ListenAndServe(addr, proxy))
}
