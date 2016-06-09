package main

import (
	"github.com/matiasinsaurralde/goproxy"
	"github.com/matiasinsaurralde/goproxy/transport"

	"github.com/googollee/go-socket.io"
	"encoding/json"

	"log"
	"flag"
	"net/http"
)

type Request struct {
	Method string
	Url string
}

var requestsChannel chan Request

func main() {

	requestsChannel := make(chan Request)

	// Socket.IO setup:
	server, err := socketio.NewServer(nil)
	if err != nil {
		panic(err)
	}

	server.On("connection", func(so socketio.Socket) {

		log.Println( "Incoming Socket.IO connection" )
		for {
			request := <- requestsChannel
			requestJson, _ := json.Marshal( &request )
			so.Emit("request", string( requestJson ))
		}

		so.On("disconnection", func() {
			log.Println("Socket.IO disconnection")
		})
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./assets")))

	go http.ListenAndServe(":8081", nil)

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

		request := Request{
			Method: req.Method,
			Url: req.URL.String(),
		}

		go func() {
			requestsChannel <- request
		}()

		return req, nil
	})
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		return resp
	})

	log.Fatal(http.ListenAndServe(*addr, proxy))
}
