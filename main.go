package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

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
	codeDB map[string]map[string]string
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

	codeDatabaseFile, err := os.Open("files/code.json")
	if err != nil {
		log.Println("opening code database file", err.Error())
	}

	jsonParser := json.NewDecoder(codeDatabaseFile)
	if err = jsonParser.Decode(&codeDB); err != nil {
		log.Println("parsing code database file", err.Error())
	}

	log.Printf("Listening on port %s", "8090")
	log.Fatal(http.ListenAndServe(":8090", router))
}
