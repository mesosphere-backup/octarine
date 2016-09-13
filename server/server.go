package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/dcos/octarine/srv"
	"github.com/dcos/octarine/util"
	"github.com/elazarl/goproxy"
)

// Server stores the server configuration
type Server struct {
	ID           string
	Verbose      bool
	CacheTimeout int
	ListenSock   string
	WriteSock    string

	port string
}

// Run starts the server
func (sv *Server) Run() error {
	srvCache := srv.New(time.Duration(sv.CacheTimeout) * time.Second)
	srvHandler := createSRVHandler(srvCache)

	proxy := goproxy.NewProxyHttpServer()
	httpProxifier := createNonProxyHandler(proxy, "http")
	proxy.NonproxyHandler = http.HandlerFunc(httpProxifier)
	proxy.OnRequest(dstSuffixMatch(util.DcosDomain)).DoFunc(stripDcosDomain)
	proxy.OnRequest(dstFirstCharMatch("_"[0])).DoFunc(srvHandler)
	proxy.Verbose = sv.Verbose

	netl, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", "0"))
	if err != nil {
		return err
	}
	_, port, err := net.SplitHostPort(netl.Addr().String())
	if err != nil {
		return err
	}
	sv.port = port
	s := &http.Server{
		Handler: proxy,
	}

	go sv.runListener()
	return s.Serve(netl)
}

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

	r.URL.Host = strings.TrimSuffix(r.URL.Host, util.DcosDomain)
	return r, nil
}

func createSRVHandler(Cache srv.Cache) func(
	r *http.Request, ctx *goproxy.ProxyCtx) (
	*http.Request, *http.Response) {

	return func(r *http.Request, ctx *goproxy.ProxyCtx) (
		*http.Request, *http.Response) {

		if host, port, err := Cache.Get(r.URL.Host); err != nil {
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

func (sv *Server) writeResponse() {
	netw, err := net.Dial("unix", sv.WriteSock)
	if err != nil {
		log.Print("dial error: ", err)
		return
	}
	defer netw.Close()
	_, err = netw.Write([]byte(sv.port))
	if err != nil {
		log.Print("write error: ", err)
		return
	}
}

func (sv *Server) runListener() {
	if err := util.RmIfExist(sv.ListenSock); err != nil {
		log.Fatal(err)
	}

	netl, err := net.Listen("unix", sv.ListenSock)
	if err != nil {
		log.Fatal("listen error: ", err)
	}
	for {
		_, err := netl.Accept()
		if err != nil {
			log.Print("accept error: ", err)
			continue
		}
		go sv.writeResponse()
	}
}
