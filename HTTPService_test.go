package goms

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Cappta/gohelpgabs"
	"github.com/Cappta/gowait"

	. "github.com/Cappta/gofixture"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	httpListenAddress = ":1337"
	httpListenPath    = "/something"
	httpMethod        = "POST"
)

func TestHTTPService(t *testing.T) {
	inputPath := "input"
	outputPath := "output"
	requiredPaths := []string{inputPath}
	handler := func(container *gohelpgabs.Container) {
		inputContainer := container.PopPath(inputPath)
		container.SetP(inputContainer.Data(), outputPath)
	}

	done := make(chan bool)
	httpService := NewHTTPService(httpListenAddress)
	httpService.Handle(httpMethod, httpListenPath, handler, requiredPaths...)
	go func() {
		fmt.Println(httpService.Run())
		done <- true
	}()

	err := gowait.AwaitTrue(func() bool { return httpService.IsRunning() }, time.Second*5)
	if err != nil {
		panic(err)
	}

	Convey("Given a HTTP server", t, func() {
		Convey("Given any string with length between 10 and 100", func() {
			anyString := AnyString(AnyIntBetween(10, 100))
			Convey("Given a json container with the string set as value", func() {
				container := gohelpgabs.New()
				container.SetP(anyString, inputPath)
				Convey("When creating a valid request", func() {
					request, err := http.NewRequest(httpMethod, fmt.Sprintf("http://localhost:1337%s", httpListenPath), strings.NewReader(container.String()))
					request.Header.Set("Content-Type", "application/json")
					Convey("Then error should be nil", func() {
						So(err, ShouldBeNil)
						Convey("Then request should not be nil", func() {
							So(request, ShouldNotBeNil)
							Convey("Given an http client", func() {
								httpClient := &http.Client{Timeout: time.Second * 5}
								Convey("When requesting", func() {
									response, err := httpClient.Do(request)
									Convey("Then error should be nil", func() {
										So(err, ShouldBeNil)
										Convey("Then response should not be nil", func() {
											So(response, ShouldNotBeNil)
											Convey("When reading response's body", func() {
												responseBody, err := ioutil.ReadAll(response.Body)
												Convey("Then error should be nil", func() {
													So(err, ShouldBeNil)
													Convey("Then data should not be nil", func() {
														So(responseBody, ShouldNotBeNil)
														Convey("Then response's body should resemble expected output", func() {
															container := gohelpgabs.New()
															container.SetP(anyString, outputPath)

															expectedOutput := container.String()
															So(string(responseBody), ShouldResemble, expectedOutput)
															Convey("Then response status should be OK", func() {
																So(response.StatusCode, ShouldEqual, http.StatusOK)
															})
														})
													})
												})
											})
										})
									})
								})
							})
						})
					})
				})
			})
		})
	})

	err = httpService.Stop()
	if err != nil {
		panic(err)
	}
}
