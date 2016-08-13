package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Route holds the configuration for a single endpoint
type Route struct {
	Method string
	URI    string
	Handle handleFuncWrapper
}

var (
	routes = []Route{
		Route{
			Method: "POST",
			URI:    "/v1/bills",
			Handle: PostBill,
		},
	}

	apiKey string
)

func init() {
	flag.StringVar(&apiKey, "api-key", "", "The Microsoft Cognitive Services API key.")
}

func main() {

	flag.Parse()

	if apiKey == "" {
		log.Fatal("API Key not provided")
	}

	router := httprouter.New()

	for _, route := range routes {
		router.Handle(route.Method, route.URI, HandleFunc(route.Handle))
	}

	log.Printf("Listening on port %s", "8090")
	log.Fatal(http.ListenAndServe(":8090", router))
}
