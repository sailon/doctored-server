package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	parseError = "Set the Content-Type header, dude."
)

type httpError struct {
	StatusCode int
	Message    string
}

type httpResponse struct {
	ContentType string
	Payload     interface{}
}

type handleFuncWrapper func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (*httpResponse, *httpError)

// HandleFunc wraps every endpoint function and handles common logic
func HandleFunc(f handleFuncWrapper) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		resp, err := f(w, r, ps)
		if err != nil {
			http.Error(w, err.Message, err.StatusCode)
		} else {
			switch ct := resp.ContentType; ct {
			case "text/plain":
				w.Header().Set("Content-Type", ct)
				fmt.Fprint(w, resp.Payload)
			default:
				w.Header().Set("Content-Type", ct)
				json.NewEncoder(w).Encode(&resp.Payload)
			}
		}
	}
}

func decodeRequestPayload(r *http.Request, value interface{}) error {
	defer r.Body.Close()

	if r.Header.Get("Content-Type") != "application/json" {
		return errors.New(parseError)
	}

	return json.NewDecoder(r.Body).Decode(&value)
}
