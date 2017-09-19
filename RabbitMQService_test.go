package goms

import (
	"fmt"
	"testing"
	"time"

	. "github.com/Cappta/gofixture"
	"github.com/Cappta/gohelpgabs"
	"github.com/Cappta/gohelprabbitmq"
	"github.com/Cappta/gowait"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	serviceQueueName = "TestForRabbitMQService"
	rpcQueueName     = "TestForRabbitMQServiceRPC"
)

func TestRabbitMQService(t *testing.T) {
	inputPath := "input"
	outputPath := "output"
	requiredPaths := []string{inputPath}
	handler := func(container *gohelpgabs.Container) {
		inputContainer := container.PopPath(inputPath)
		container.SetP(inputContainer.Data(), outputPath)
	}

	rabbitMQ := gohelprabbitmq.ConnectLocallyWithDefaultUser()
	rpc := gohelprabbitmq.NewRPC(gohelprabbitmq.NewSimpleConsumer(rabbitMQ, rpcQueueName), time.Second*5)
	publisher := gohelprabbitmq.NewSimplePublisher(rabbitMQ, fmt.Sprintf("@%s", serviceQueueName))

	done := make(chan bool)
	rabbitMQService := NewRabbitMQService(rabbitMQ)
	go func() {
		fmt.Println(rabbitMQService.Consume(serviceQueueName, handler, requiredPaths...))
		done <- true
	}()
	go rpc.Consume()

	err := gowait.AwaitNotNil(func() interface{} { return rabbitMQService.GetConsumer(serviceQueueName) }, time.Second*5)
	if err != nil {
		panic(err)
	}
	err = gohelprabbitmq.NewQueueObserver(rabbitMQ, rabbitMQService.GetConsumer(serviceQueueName).QueueSettings).AwaitConsumer(time.Second * 5)
	if err != nil {
		panic(err)
	}
	err = gohelprabbitmq.NewQueueObserver(rabbitMQ, rpc.QueueSettings).AwaitConsumer(time.Second * 5)
	if err != nil {
		panic(err)
	}

	Convey("Given a running async consumer", t, func() {
		Convey("Given any string with length between 10 and 100", func() {
			anyString := AnyString(AnyIntBetween(10, 100))
			Convey("Given a json container with the string set as value", func() {
				container := gohelpgabs.New()
				container.SetP(anyString, inputPath)
				Convey("When preparing the rpc", func() {
					callback := rpc.Prepare(container)
					Convey("Then callback should not be nil", func() {
						So(callback, ShouldNotBeNil)
						Convey("When publishing to the test service", func() {
							publisher.Publish(container.Bytes())
							Convey("Then rpc should receive the message within 5 seconds", func() {
								select {
								case <-time.After(time.Second * 5):
									t.Fail()
								case returnedMessage := <-callback:
									Convey("Then message should not be nil", func() {
										So(returnedMessage, ShouldNotBeNil)
										Convey("Then message's body should resemble expected output", func() {
											container := gohelpgabs.New()
											container.SetP(anyString, outputPath)

											expectedOutput := container.String()
											So(string(returnedMessage.Body), ShouldResemble, expectedOutput)
										})
									})
								}
							})
						})
					})
				})
			})
		})
	})

	err = rabbitMQService.StopConsuming(serviceQueueName)
	if err != nil {
		panic(err)
	}
	err = rpc.StopConsuming()
	if err != nil {
		panic(err)
	}
	<-done
}