package main

import (
	"github.com/matiasinsaurralde/goproxy"
	"github.com/matiasinsaurralde/goproxy/transport"

	"github.com/googollee/go-socket.io"

	"log"
	"flag"
	"net/http"
)

func main() {

	// Socket.IO setup:



	// Proxy setup:


	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8080", "proxy listen address")
	flag.Parse()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	tr := transport.Transport{Proxy: transport.ProxyFromEnvironment}

	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
			ctx.UserData, resp, err = tr.DetailedRoundTrip(req)
			return
		})
		log.Println(req, ctx)
		// logger.LogReq(req, ctx)
		return req, nil
	})
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		// logger.LogResp(resp, ctx)
		log.Println(resp, ctx)
		return resp
	})

	log.Fatal(http.ListenAndServe(*addr, proxy))
}
