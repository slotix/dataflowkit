package main

import (
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
	"fmt"
	"github.com/gorilla/mux"

	"net/http"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
)

// Handler is the exported handler called by AWS Lambda.
var Handler apigatewayproxy.Handler

func init() {
	Handler = NewHandler()
}

func NewHandler() apigatewayproxy.Handler {
	ln := net.Listen()

	// Amazon API Gateway Binary support out of the box.
	handle := apigatewayproxy.New(ln, nil).Handle

	// Any Go framework complying with the Go http.Handler interface can be used.
	// This includes, but is not limited to, Vanilla Go, Gin, Echo, Gorrila, Goa, etc.
	go http.Serve(ln, setUpMux())

	return handle
}

func setUpMux() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/todos", create).Methods(http.MethodPost)
	r.HandleFunc("/todos", list).Methods(http.MethodGet)
	r.HandleFunc("/todos/{id}", read).Methods(http.MethodGet)
	r.HandleFunc("/todos/{id}", update).Methods(http.MethodPut)
	r.HandleFunc("/todos/{id}", remove).Methods(http.MethodDelete)

	return r
}

func create(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("X-Powered-By", "serverless-golang")
	fmt.Fprintf(w, "[%d] Created", http.StatusCreated)
}

func read(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	w.WriteHeader(http.StatusOK)
	w.Header().Set("X-Powered-By", "serverless-golang")
	fmt.Fprintf(w, "%d - Reading Id: %s", http.StatusOK, id)
}

func list(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("X-Powered-By", "serverless-golang")
	fmt.Fprintf(w, "%d - Listing All", http.StatusOK)
}

func update(w http.ResponseWriter, r *http.Request) {
	id:= mux.Vars(r)["id"]

	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("X-Powered-By", "serverless-golang")
	fmt.Fprintf(w, "%d - Updated Id: %s", http.StatusNoContent, id)
}

func remove(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("X-Powered-By", "serverless-golang")
	fmt.Fprintf(w, "%d - Deleted Id: %s", http.StatusNoContent, id)
}
