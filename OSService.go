package goms

import (
	"flag"
	"log"

	"github.com/kardianos/service"
)

type action func() error

// OSService is a Operational System Service helper to setup a working service
type OSService struct {
	service service.Service

	start action
	stop  action
}

// NewOSService creates a new OSService
func NewOSService(shortName, longName, description string, start, stop action) (*OSService, error) {
	config := &service.Config{
		Name:        shortName,
		DisplayName: longName,
		Description: description,
	}

	osService := &OSService{
		start: start,
		stop:  stop,
	}

	service, err := service.New(osService, config)
	if err != nil {
		return nil, err
	}

	osService.service = service
	return osService, nil
}

// Main is the function you're supposed to call to run the service and support CLI
func (osService *OSService) Main() (err error) {
	installFlag := flag.Bool("Install", false, "Use install to install the service")
	uninstallFlag := flag.Bool("Uninstall", false, "Use Uninstall to uninstall the service")
	flag.Parse()

	if installFlag != nil && *installFlag == true {
		if err = osService.service.Install(); err != nil {
			log.Println("Failed to install service: ", err)
			return
		}
		log.Println("Installed service")
		if err = osService.service.Start(); err != nil {
			log.Println("Failed to start service: ", err)
			return
		}
		log.Println("Started service")
		return
	}

	if uninstallFlag != nil && *uninstallFlag == true {
		if err = osService.service.Stop(); err != nil {
			log.Println("Failed to stop service: ", err)
		}
		log.Println("Stopped service")
		if err = osService.service.Uninstall(); err != nil {
			log.Println("Failed to uninstall service: ", err)
			return
		}
		log.Println("Uninstalled service")
		return
	}

	log.Println("Running service")
	defer log.Println("Service stopped")
	return osService.service.Run()
}

// Start provides a place to initiate the service. The service doesn't not
// signal a completed start until after this function returns, so the
// Start function must not take more then a few seconds at most.
func (osService *OSService) Start(s service.Service) error {
	return osService.start()
}

// Stop provides a place to clean up program execution before it is terminated.
// It should not take more then a few seconds to execute.
// Stop should not call os.Exit directly in the function.
func (osService *OSService) Stop(s service.Service) error {
	return osService.stop()
}
