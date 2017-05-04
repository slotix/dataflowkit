package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/slotix/dataflowkit/splash"
)

func makeFetchEndpoint(svc ParseService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(splash.Request)
		v, err := svc.Fetch(req)
		//v, err := svc.GetHTML(request.(string))
		if err != nil {
			//	return getHTMLResponse{v, err.Error()}, nil
			//			return errResponse{err.Error()}, nil
			return nil, err
		}
		//return getHTMLResponse{v, ""}, nil
		return v, nil
	}
}

func makeMarshalDataEndpoint(svc ParseService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		//	fmt.Println("from makeMarshalDataEndpoint",string(request.([]byte)))
		v, err := svc.ParseData(request.([]byte))
		if err != nil {
			//return errResponse{err.Error()}, nil
			return nil, err
		}
		return v, nil
	}
}

/*
func makeCheckServicesEndpoint(svc ParseService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		status := svc.CheckServices()
		return status, nil
	}
}
*/

func decodeFetchRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request splash.Request
	//request, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//	fmt.Println(err)
	//}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeMarshalDataRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request, err := ioutil.ReadAll(r.Body)
	//fmt.Println("from decodeMarshalDataRequest",string(request))
	if err != nil {
		//fmt.Println(err)
		return nil, err
	}

	return request, nil
}

func decodeCheckServicesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		//	fmt.Println(err)
		return nil, err
	}
	return request, nil
}

func decodeFetchResponse(_ context.Context, r *http.Response) (interface{}, error) {
	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		//	fmt.Println(err)
		return nil, err
	}
	return string(response), nil
}

func decodeMarshalDataResponse(_ context.Context, r *http.Response) (interface{}, error) {
	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func decodeCheckServicesResponse(_ context.Context, r *http.Response) (interface{}, error) {
	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		//	fmt.Println(err)
		return nil, err
	}
	return string(response), nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	//_, err := w.Write(response.([]byte))
	data, err := ioutil.ReadAll(response.(io.Reader))
	if err != nil {
		return err
	}
	_, err = w.Write(data)

	if err != nil {
		return err
	}
	return nil
}

func encodeCheckServicesResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func encodeRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf *bytes.Buffer
	buf = bytes.NewBuffer(request.([]byte))
	r.Body = ioutil.NopCloser(buf)
	return nil
}

type checkServicesResponse struct {
	Status map[string]string `json:"status,omitempty"`
}
