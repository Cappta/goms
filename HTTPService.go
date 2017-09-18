package goms

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Cappta/gohelpgabs"
	"github.com/gorilla/mux"
)

var (
	errServiceRunning    = fmt.Errorf("Service is already running")
	errServiceNotRunning = fmt.Errorf("Service is not running")
)

// HTTPService is a framework for HTTP services
type HTTPService struct {
	router  *mux.Router
	server  *http.Server
	address string

	running bool
}

// NewHTTPService creates a new HTTPService given a listen address
func NewHTTPService(address string) (service *HTTPService) {
	if address == "" {
		address = ":0" //Listen on any port
	}

	service = &HTTPService{}
	service.router = mux.NewRouter()
	service.address = address
	return
}

// IsRunning returns true if the service is running
func (service *HTTPService) IsRunning() bool {
	return service.running
}

// Handle appends an entry point for the HTTP service given the callback and required paths
func (service *HTTPService) Handle(method, path string, handle handler, requiredPaths ...string) {
	service.router.Methods(method).Path(path).HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			defer func() {
				err := recover()
				if err == nil {
					return
				}

				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "PANIC: %v", err)
			}()

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("Failed to read request body: ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			container, err := gohelpgabs.ParseJSON(body)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Error parsing JSON \"%s\" Content: %s", err.Error(), string(body))
				return
			}

			if missingPaths := container.GetMissingPaths(requiredPaths...); len(missingPaths) > 0 {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Message did not contain: %s", strings.Join(missingPaths, ", "))
				return
			}

			handle(container)
			fmt.Fprintf(w, container.String())
		},
	)
}

// Run starts the service
func (service *HTTPService) Run() error {
	if service.running == true {
		return errServiceRunning
	}

	service.running = true
	service.server = &http.Server{Addr: service.address, Handler: service.router}
	err := service.server.ListenAndServe()
	service.running = false
	return err
}

// Stop stops the service
func (service *HTTPService) Stop() error {
	if service.running == false {
		return errServiceNotRunning
	}

	return service.server.Close()
}
