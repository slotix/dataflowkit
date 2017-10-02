package main

import (
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
	"net/http"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"fmt"
	"encoding/json"
	"time"
	"log"
	"strings"
)

var Handler apigatewayproxy.Handler

type ResponseToSlack struct {
	ResponseType string       `json:"response_type"`
	Text         string       `json:"text"`
}

type DogResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func init() {
	Handler = NewHandler()
}

func NewHandler() apigatewayproxy.Handler {
	ln := net.Listen()

	handle := apigatewayproxy.New(ln, nil).Handle

	http.HandleFunc("/dogs", handleDogs)

	go http.Serve(ln, nil)

	return handle
}

func handleDogs(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form.", http.StatusBadRequest)
		return
	}

	breed := r.Form.Get("text")

	url := fmt.Sprintf("https://dog.ceo/api/breed/%s/images/random", formatDogName(breed))

	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	dogResponse := DogResponse{}

	if err := json.NewDecoder(res.Body).Decode(&dogResponse); err != nil {
		log.Println("error when parsing the response: ", err)
	}

	sr := ResponseToSlack{ResponseType: "in_channel", Text: dogResponse.Message}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sr)
}

func formatDogName(s string) string {
	words := strings.Fields(s)
	if len(words) == 2 {
		for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
			words[i], words[j] = words[j], words[i]
		}

		return strings.Join(words, "/")
	}
	return s
}
