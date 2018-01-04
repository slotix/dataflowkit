package fetch

import (
	"net"
	"net/http"
)
const addr = "localhost:12345"

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Conent-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`))
}

func init() {	
	server := &http.Server{}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", indexHandler)
	go func() {
		if err := server.Serve(listener); err != nil {
			logger.Error("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
}
