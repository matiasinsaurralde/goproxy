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
	Proto string `json:"protocol"`
	Method string `json:"method"`
	Url string	`json:"url"`
}

var requestsChannel chan Request

func main() {

	requestsChannel := make(chan Request)

	// Socket.IO setup:
	server, err := socketio.NewServer(nil)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			request := <- requestsChannel
			log.Println("Emiting", request)

			requestJson, _ := json.Marshal( &request )

			server.BroadcastTo("requests", "request", string(requestJson))
		}
	}()

	server.On("connection", func(so socketio.Socket) {
		so.Join("requests")
		log.Println( "Incoming Socket.IO connection" )

		so.On("disconnection", func() {
			log.Println("Socket.IO disconnection")
		})
	})

	http.HandleFunc("/socket.io/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		origin := req.Header.Get("Origin")
		w.Header().Add("Access-Control-Allow-Origin", origin)

		server.ServeHTTP(w,req)
	})
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
			Proto: req.Proto,
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
