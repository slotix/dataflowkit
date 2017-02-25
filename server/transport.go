package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"fmt"

	"context"

	"github.com/go-kit/kit/endpoint"
)

func makeGetHTMLEndpoint(svc ParseService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getHTMLRequest)
		v, err := svc.GetHTML(req.URL)
		fmt.Println("from makeGetHTMLEndpoint",request)
		//v, err := svc.GetHTML(request.(string))

		if err != nil {
			//	return getHTMLResponse{v, err.Error()}, nil
			return errResponse{err.Error()}, nil

		}
		//return getHTMLResponse{v, ""}, nil
		return v, nil
	}
}

func makeMarshalDataEndpoint(svc ParseService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		//	fmt.Println("from makeMarshalDataEndpoint",string(request.([]byte)))
		v, err := svc.MarshalData(request.([]byte))
		if err != nil {
			//return marshalDataResponse{v, err.Error()}, nil
			return errResponse{err.Error()}, nil
		}
		return v, nil
		//return marshalDataResponse{v, ""}, nil
	}
}

func makeCheckServicesEndpoint(svc ParseService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		status := svc.CheckServices()
		return status, nil
	}
}

func decodeGetHTMLRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request getHTMLRequest
	//request, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//	fmt.Println(err)
	//}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	//fmt.Println(string(request))
	return request, nil
}

func decodeMarshalDataRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request, err := ioutil.ReadAll(r.Body)
	//fmt.Println("from decodeMarshalDataRequest",string(request))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//fmt.Println("from decodeMarshalDataRequest",string(request))
	//var request marshalDataRequest

	//if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
	//	return nil, err
	//}

	return request, nil
}

func decodeCheckServicesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return request, nil
}

func decodeGetHTMLResponse(_ context.Context, r *http.Response) (interface{}, error) {
	//var response getHTMLResponse

	//if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
	//	return nil, err
	//}
	//return response, nil
	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return string(response), nil
}

func decodeMarshalDataResponse(_ context.Context, r *http.Response) (interface{}, error) {
	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return string(response), nil
	//var response marshalDataResponse
	//if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
	//	return nil, err
	//}
	///return response, nil
}

func decodeCheckServicesResponse(_ context.Context, r *http.Response) (interface{}, error) {
	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return string(response), nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	//err := json.NewEncoder(w).Encode(response)
	return json.NewEncoder(w).Encode(response)
	//return err
}

/*
func encodeRequest(_ context.Context, r *http.Request, request interface{}) error {
	//fmt.Println("from encodeRequest", string(request.([]uint8)))
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	fmt.Println("from encodeRequest", r)

	return nil
}
*/

func encodeRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf *bytes.Buffer
	buf = bytes.NewBuffer(request.([]byte))
	r.Body = ioutil.NopCloser(buf)
	return nil
}

type errResponse struct {
	Err string `json:"err,omitempty"`
}

type getHTMLRequest struct {
	URL string `json:"url"`
}



type checkServicesResponse struct {
	Status map[string]string `json:"status,omitempty"`
}
