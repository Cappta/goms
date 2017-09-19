package goms

import (
	"fmt"
	"testing"
	"time"

	"github.com/Cappta/gohelpgabs"
	"github.com/Cappta/gohelprabbitmq"
	"github.com/Cappta/gowait"

	. "github.com/Cappta/debugo"
	. "github.com/Cappta/gofixture"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	serviceQueueName = "TestForRabbitMQService"
	rpcQueueName     = "TestForRabbitMQServiceRPC"
)

func TestRabbitMQService(t *testing.T) {
	rabbitMQ := gohelprabbitmq.ConnectLocallyWithDefaultUser()
	rpc := gohelprabbitmq.NewRPC(gohelprabbitmq.NewSimpleConsumer(rabbitMQ, rpcQueueName), time.Second*5)
	publisher := gohelprabbitmq.NewSimplePublisher(rabbitMQ, fmt.Sprintf("@%s", serviceQueueName))

	done := make(chan bool)
	rabbitMQService := NewRabbitMQService(rabbitMQ)
	go func() {
		fmt.Println(rabbitMQService.Consume(serviceQueueName, inputOutputHandler, requiredPaths...))
		done <- true
	}()
	go rpc.Consume()

	PanicOnError(gowait.AwaitNotNil(func() interface{} { return rabbitMQService.GetConsumer(serviceQueueName) }, time.Second*5))
	PanicOnError(gohelprabbitmq.NewQueueObserver(rabbitMQ, rabbitMQService.GetConsumer(serviceQueueName).QueueSettings).AwaitConsumer(time.Second * 5))
	PanicOnError(gohelprabbitmq.NewQueueObserver(rabbitMQ, rpc.QueueSettings).AwaitConsumer(time.Second * 5))

	Convey("Given a running async consumer", t, func() {
		anyString := AnyString(AnyIntBetween(10, 100))
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
								Convey("And message's body should resemble expected output", func() {
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

	PanicOnError(rabbitMQService.StopConsuming(serviceQueueName))
	PanicOnError(rpc.StopConsuming())
	<-done
}
