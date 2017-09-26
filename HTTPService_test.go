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

	. "github.com/Cappta/debugo"
	. "github.com/Cappta/gofixture"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	httpListenAddress = ":1337"
	httpListenPath    = "/something"
	httpMethod        = "POST"
)

func TestHTTPService(t *testing.T) {
	done := make(chan bool)
	httpService := NewHTTPService(httpListenAddress)
	httpService.Handle(httpMethod, httpListenPath, inputOutputHandler)
	go func() {
		fmt.Println(httpService.Run())
		done <- true
	}()

	PanicOnError(gowait.AwaitTrue(func() bool { return httpService.IsRunning() }, time.Second*5))

	Convey("Given a request to the server", t, func() {
		anyString := AnyString(AnyIntBetween(10, 100))
		container := gohelpgabs.New()
		container.SetP(anyString, inputPath)

		request, err := http.NewRequest(httpMethod, fmt.Sprintf("http://localhost:1337%s", httpListenPath), strings.NewReader(container.String()))
		PanicOnError(err)
		request.Header.Set("Content-Type", "application/json")

		httpClient := &http.Client{Timeout: time.Second * 5}
		response, err := httpClient.Do(request)
		PanicOnError(err)

		Convey("Then response should not be nil", func() {
			So(response, ShouldNotBeNil)

			Convey("When reading response's body", func() {
				responseBody, err := ioutil.ReadAll(response.Body)
				PanicOnError(err)

				Convey("Then data should not be nil", func() {
					So(responseBody, ShouldNotBeNil)

					Convey("And response's body should resemble expected output", func() {
						container := gohelpgabs.New()
						container.SetP(anyString, outputPath)

						expectedOutput := container.String()
						So(string(responseBody), ShouldResemble, expectedOutput)
						Convey("And response status should be OK", func() {
							So(response.StatusCode, ShouldEqual, http.StatusOK)
						})
					})
				})
			})
		})
	})

	PanicOnError(httpService.Stop())
	<-done
}
