package goms

import (
	"flag"

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
		err = osService.service.Install()
		if err != nil {
			return
		}
		return osService.service.Start()
	}

	if uninstallFlag != nil && *uninstallFlag == true {
		osService.service.Stop()
		return osService.service.Uninstall()
	}

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
