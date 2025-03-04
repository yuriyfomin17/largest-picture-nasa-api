package pkg

import (
	"errors"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

const RabbitmqQueue = "command-queue"
const RabbitmqExchange = "command-exchange"
const RabbitmqRoutingKey = "command-routing-key"

func ConnectRabbitMQ(rabbitMQURL string) (*RabbitMQ, error) {
	// Establish a connection to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, err
	}

	// Open a channel on the connection
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declare the exchange
	err = channel.ExchangeDeclare(
		RabbitmqExchange, // Name
		"direct",         // Type (options: "direct")
		true,             // Durable
		false,            // Auto-delete when unused
		false,            // Internal
		false,            // No-wait
		nil,              // Arguments
	)
	if err != nil {
		return nil, errors.New("could not declare exchange")
	}

	_, err = channel.QueueDeclare(
		RabbitmqQueue, // Name
		true,          // Durable
		false,         // Auto-delete when unused
		false,         // Exclusive
		false,         // No-wait
		nil,           // Arguments
	)
	if err != nil {
		return nil, errors.New("could not declare queue")
	}

	// Bind the queue to the exchange with a routing key
	err = channel.QueueBind(
		RabbitmqQueue,      // Queue name
		RabbitmqRoutingKey, // Routing key
		RabbitmqExchange,   // Exchange name
		false,              // No-wait
		nil,                // Arguments
	)
	if err != nil {
		return nil, errors.New("could not bind queue to exchange")
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
	}, nil
}

func (r *RabbitMQ) PublishCommand(solCommand int) error {
	byteArrayCommand := []byte(strconv.Itoa(solCommand))
	err := r.channel.Publish(
		RabbitmqExchange,   // Exchange
		RabbitmqRoutingKey, // Routing key (queue)
		false,              // Mandatory
		false,              // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        byteArrayCommand,
		},
	)
	return err
}

// GetMessage messages from the queue
func (r *RabbitMQ) GetMessage() <-chan amqp.Delivery {
	messages, err := r.channel.Consume(
		RabbitmqQueue, // Queue name
		"",            // Consumer name
		true,          // Auto-acknowledge
		false,         // Exclusive
		false,         // No-local
		false,         // No-wait
		nil,           // Arguments
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to consume messages from RabbitMQ")
	}
	return messages
}

// Close RabbitMQ connection and channel
func (r *RabbitMQ) Close() {
	err := r.channel.Close()
	if err != nil {
		return
	}
	err = r.conn.Close()
	if err != nil {
		return
	}
}
