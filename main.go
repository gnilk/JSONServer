package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var fileToSend string = "data.json"
var endpoint string = "data"
var servicePort int = 3080

func initialize() error {
	return nil
}

func main() {
	start()
}

func start() {

	log.Printf("Initialize")
	err := initialize()
	if err != nil {
		log.Fatalf("Initialization failed, error: %s\n", err.Error())
	}
	serviceIP := "0.0.0.0"
	serviceAddress := fmt.Sprintf("%v:%v", serviceIP, servicePort)
	log.Printf("Starting service at: %v\n", serviceAddress)
	log.Printf("  serving GET requests from endpoint: %s, sending: %s\n", path.Join("/", endpoint), fileToSend)
	log.Printf("  serving POST requests to endpoint: %s\n", path.Join("/", endpoint))

	r, err := NewHTTPApi()
	if err != nil {
		log.Panicln(err)
	}

	// Where ORIGIN_ALLOWED is like `scheme://dns[:port]`, or `*` (insecure)

	err = http.ListenAndServe(serviceAddress, handlers.LoggingHandler(os.Stdout, r))

	if err != nil {
		log.Panicln(err)
	}
}

//
// APIServer is the middle-ware interface
//
type APIServer struct {
	Handler func(http.ResponseWriter, *http.Request) (int, error)
}

//
// NewHTTPApi constructor
//
func NewHTTPApi() (*mux.Router, error) {
	r := mux.NewRouter()

	serverEndpoint := path.Join("/", endpoint)

	r.Handle(serverEndpoint, APIServer{getHandler}).Methods("GET")
	r.Handle(serverEndpoint, APIServer{postHandler}).Methods("POST")
	return r, nil
}

//
// ServeHTTP is just the wrapper which calls the handler for the API
//
func (ah APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.Handler(w, r)
}

func sendRawAsJSONResponse(w http.ResponseWriter, payload []byte) (int, error) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(payload))

	return http.StatusOK, nil
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) (int, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return sendRawAsJSONResponse(w, payload)
}

func getHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	// import data
	data, err := ioutil.ReadFile(fileToSend)
	if err != nil {
		log.Fatalf("Unable to read file: %s\n", fileToSend)
		return http.StatusInternalServerError, err
	}
	log.Printf("Sending:\n%s\n", data)
	return sendRawAsJSONResponse(w, data)
}

func dumpMap(space string, m map[string]interface{}) {
	for k, v := range m {
		if mv, ok := v.(map[string]interface{}); ok {
			fmt.Printf("{ \"%v\": \n", k)
			dumpMap(space+"\t", mv)
			fmt.Printf("}\n")
		} else {
			fmt.Printf("%v %v : %v\n", space, k, v)
		}
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	// import data
	log.Printf("got request, dumping body\n")
	data, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return http.StatusInternalServerError, err
	}
	log.Printf("Body: \n%s\n", string(data))

	log.Printf("Parsing JSON\n")
	jsonMap := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&jsonMap)
	if err != nil {
		log.Printf("Failed to unmarshal: %v\n", err)
	} else {
		log.Printf("Ok, dumping JSON data\n")
		dumpMap("", jsonMap)
	}

	return http.StatusOK, nil
}
