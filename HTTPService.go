package goms

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Cappta/gohelpgabs"
	"github.com/gorilla/mux"
)

var (
	httpAddrKey = "HTTP_ListenAddress"

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
func NewHTTPService(env map[string]string) (service *HTTPService) {
	address := env[httpAddrKey]
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
func (service *HTTPService) Handle(method, path string, handle handler) {
	service.router.Methods(method).Path(path).HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			defer service.recover(w)

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fprintfAndLog(w, "Failed to read request body: ", err)
				return
			}
			container, err := gohelpgabs.ParseJSON(body)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fprintfAndLog(w, "Error parsing JSON \"%s\" Content: %s", err.Error(), string(body))
				return
			}

			handle(container)
			fmt.Fprintf(w, container.String())
		},
	)
}

func (service *HTTPService) recover(w http.ResponseWriter) {
	err := recover()
	if err == nil {
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	fprintfAndLog(w, "PANIC: %v", err)
}

func fprintfAndLog(w io.Writer, format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	fmt.Fprintf(w, message)
	log.Println(message)
}

// Run starts the service
func (service *HTTPService) Run() (err error) {
	if service.running {
		return errServiceRunning
	}

	service.running = true
	for {
		service.server = &http.Server{Addr: service.address, Handler: service.router}
		err = service.server.ListenAndServe()
		if service.running == false {
			return
		}
		log.Println("HTTPService restarting: ", err)
	}
}

// Stop stops the service
func (service *HTTPService) Stop() error {
	if service.running == false {
		return errServiceNotRunning
	}

	service.running = false
	return service.server.Close()
}
