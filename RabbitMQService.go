package goms

import (
	"fmt"
	"log"
	"strings"

	"github.com/Cappta/gohelpgabs"
	"github.com/Cappta/gohelprabbitmq"
	"github.com/streadway/amqp"
)

var (
	errConsumerNotFound = fmt.Errorf("The specified consumer was not found")
)

// RabbitMQService is a framework for RabbitMQ services
type RabbitMQService struct {
	connection  *gohelprabbitmq.Connection
	consumerMap map[string]*gohelprabbitmq.SimpleConsumer
}

// NewRabbitMQService creates a new RabbitMQService given a RabbitMQ connection
func NewRabbitMQService(connection *gohelprabbitmq.Connection) *RabbitMQService {
	return &RabbitMQService{
		connection:  connection,
		consumerMap: make(map[string]*gohelprabbitmq.SimpleConsumer),
	}
}

// Consume will consume the given queue, asserting the required paths, pass data to be handled and route the response
func (service *RabbitMQService) Consume(queueName string, handle handler, requiredPaths ...string) (err error) {
	forwardToPath := fmt.Sprintf("%s.ForwardTo", queueName)
	requiredPaths = append(requiredPaths, forwardToPath)

	router := gohelprabbitmq.NewMessageRouter(forwardToPath, service.connection)
	consumer := gohelprabbitmq.NewSimpleConsumer(service.connection, queueName)
	service.consumerMap[queueName] = consumer

	return consumer.Consume(func(delivery amqp.Delivery) {
		container, err := gohelpgabs.ParseJSON(delivery.Body)

		if err != nil {
			log.Println("Error parsing JSON \"", err, "\" Content: ", string(delivery.Body))
			delivery.Reject(false)
			return
		}

		if missingPaths := container.GetMissingPaths(requiredPaths...); len(missingPaths) > 0 {
			log.Println("Message did not contain : " + strings.Join(missingPaths, ", "))
			delivery.Reject(false)
			return
		}

		handle(container)

		err = router.Route(container)
		if err != nil {
			fmt.Println("Failed to respond : " + err.Error())
		}

		err = delivery.Ack(false)
		if err != nil {
			fmt.Println("Failed to ack a message : " + err.Error())
		}
	})
}

// StopConsuming will stop consuming the provided queue
func (service *RabbitMQService) StopConsuming(queueName string) (err error) {
	consumer, found := service.consumerMap[queueName]
	if found == false {
		return errConsumerNotFound
	}

	return consumer.StopConsuming()
}

// GetConsumer retrieves the consumer for the provided queue
func (service *RabbitMQService) GetConsumer(queueName string) (consumer *gohelprabbitmq.SimpleConsumer) {
	return service.consumerMap[queueName]
}
