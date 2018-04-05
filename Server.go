package goms

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Cappta/debugo"
	"github.com/Cappta/gowait"
	"github.com/joho/godotenv"
)

var (
	gomsPath = os.ExpandEnv("${PROGRAMDATA}\\GoMS")

	defaultSettings = `HTTP_ListenAddress=":80"`
)

type Server struct {
	osService *OSService

	httpService *HTTPService

	env map[string]string

	running bool
}

func NewServer(shortName, longName, description string) (server *Server) {
	servicePath := gomsPath + "\\" + shortName
	logPath := servicePath + "\\Log"
	debugo.PanicOnError(os.MkdirAll(logPath, 0666))

	logFile, err := os.Create(logPath + time.Now().Format("\\2006_01_02_15_04_05.0000000.txt"))
	debugo.PanicOnError(err)
	log.SetOutput(logFile)

	server = &Server{}
	debugo.PanicOnError(server.readEnvFile(servicePath + "\\Settings.env"))

	server.osService, err = NewOSService(shortName, longName, description, server.onServiceStart, server.onServiceStop)
	debugo.PanicOnError(err)
	return
}

func (server *Server) readEnvFile(envFilePath string) (err error) {
	if _, err = os.Stat(envFilePath); os.IsNotExist(err) {
		var file *os.File
		if file, err = os.Create(envFilePath); err != nil {
			return
		}

		if _, err = fmt.Fprint(file, defaultSettings); err != nil {
			return
		}
		err = file.Close()
	}
	if err != nil {
		return
	}

	server.env, err = godotenv.Read(envFilePath)
	return
}

func (server *Server) HandleHTTP(method, path string, handle handler) {
	if server.httpService == nil {
		server.httpService = NewHTTPService(server.env)
	}

	server.httpService.Handle(method, path, handle)
}

func (server *Server) Main() {
	panic(server.osService.Main())
}

func (server *Server) onServiceStart() (err error) {
	defer debugo.LogPanic()

	if server.httpService != nil {
		go server.httpService.Run()

		if err = gowait.AwaitTrue(func() bool { return server.httpService.IsRunning() }, time.Second); err != nil {
			return err
		}
	}

	return
}

func (server *Server) onServiceStop() (err error) {
	defer debugo.LogPanic()

	if server.httpService != nil && server.httpService.IsRunning() {
		if err = server.httpService.Stop(); err != nil {
			return
		}
	}
	return
}
