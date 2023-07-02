package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"logs_topic", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments

	)
}

func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	// We are returning a queue with these attributes
	return ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when un-used
		true,  // exclusive
		false, // no-wait
		nil,   // arguments

	)
}
