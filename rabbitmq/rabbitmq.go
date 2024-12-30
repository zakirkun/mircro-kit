package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Address string
}

// Open implements IRabbitMQ.
func (r RabbitMQ) Open() (*amqp.Connection, error) {
	conn, err := amqp.Dial(r.Address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn, nil
}

// Publish implements IRabbitMQ.
func (r RabbitMQ) Publish(exchangeName, routingKey string, body interface{}) error {
	conn, err := amqp.Dial(r.Address)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	// Marshal the message body into JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal message body: %w", err)
	}

	// Publish the message to the specified exchange
	err = ch.Publish(
		exchangeName, // exchange name
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json", // Set ContentType to JSON
			Body:        jsonBody,           // Message body
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf(" [x] Sent message to exchange '%s' with routing key '%s': %s", exchangeName, routingKey, jsonBody)
	return nil
}

// DeclareExchange declares an exchange with the given name and type.
func (r RabbitMQ) DeclareExchange(exchangeName, exchangeType string, durable, autoDelete bool) error {
	conn, err := amqp.Dial(r.Address)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	// Declare the exchange
	err = ch.ExchangeDeclare(
		exchangeName, // name of the exchange
		exchangeType, // type of the exchange (direct, fanout, topic, headers)
		durable,      // durable
		autoDelete,   // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Printf("Exchange declared: %s (type: %s)", exchangeName, exchangeType)
	return nil
}

// Listener implements IRabbitMQ.
func (r RabbitMQ) Listener(exchangeName, queueName, routingKey string, cb ...func(payload []byte) error) {
	conn, err := amqp.Dial(r.Address)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	log.Printf("Connected to RabbitMQ for exchange: %s, queue: %s", exchangeName, queueName)

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare the exchange
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"direct",     // type (can be made configurable)
		true,         // durable
		false,        // auto-delete
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, fmt.Sprintf("Failed to declare exchange: %s", exchangeName))

	// Declare the queue
	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, fmt.Sprintf("Failed to declare queue: %s", queueName))

	// Bind the queue to the exchange with the routing key
	err = ch.QueueBind(
		q.Name,       // queue name
		routingKey,   // routing key
		exchangeName, // exchange name
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, fmt.Sprintf("Failed to bind queue: %s to exchange: %s with routing key: %s", queueName, exchangeName, routingKey))

	// Consume messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, fmt.Sprintf("Failed to register a consumer for queue: %s", queueName))

	log.Printf("Waiting for messages on exchange: %s, queue: %s", exchangeName, queueName)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message on queue: %s, Message: %s", queueName, d.Body)
			for _, f := range cb {
				err := f(d.Body)
				if err != nil {
					log.Printf("Callback RabbitMQ Failed for queue: %s, Error: %v", queueName, err)
				}
			}
		}
	}()

	<-forever
}

type IRabbitMQ interface {
	Open() (*amqp.Connection, error)
	Listener(exchangeName, queueName, routingKey string, cb ...func(payload []byte) error)
	Publish(exchangeName, routingKey string, body interface{}) error
	DeclareExchange(exchangeName, exchangeType string, durable, autoDelete bool) error
}

func NewRabbitMQ(addr string) IRabbitMQ {
	return RabbitMQ{
		Address: addr,
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}
