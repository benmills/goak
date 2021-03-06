package server

import (
	"github.com/bmizerany/pat"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"goak/httpclient"
	"goak/peer"
)

type Server struct {
	*peer.Peer
	values map[string]string
	logger *log.Logger
}

func New(url string, logger *log.Logger) *Server {
	peer := peer.New(url, logger)
	return &Server{peer, make(map[string]string), logger}
}

func (server *Server) Handler() http.Handler {
	m := pat.New()

	server.Peer.Handler(m)

	m.Put("/data/:key", http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get(":key")
		body, _ := ioutil.ReadAll(request.Body)
		value := string(body)

		destinationAddress := server.PeerAddressForKey(key)

		if server.URL() == destinationAddress {
			server.logger.Printf("Storing '%s'->'%s'", key, value)
			server.values[key] = value
			w.WriteHeader(201)
			io.WriteString(w, value)
		} else {
			server.logger.Printf("Passing off '%s'->'%s' to %s", key, value, destinationAddress)
			statusCode, response := httpclient.Put(destinationAddress+"/data/"+key, value)
			w.WriteHeader(statusCode)
			io.WriteString(w, response)
		}
	}))

	m.Get("/data/:key", http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get(":key")

		destinationAddress := server.PeerAddressForKey(key)

		if server.URL() == destinationAddress {
			value, ok := server.values[key]
			if ok {
				w.WriteHeader(200)
				io.WriteString(w, value)
			} else {
				w.WriteHeader(404)
			}
		} else {
			statusCode, response := httpclient.Get(destinationAddress+"/data/"+key, "")
			w.WriteHeader(statusCode)
			io.WriteString(w, response)
		}
	}))

	return m
}
