// initializers/rabbitmq.go
package initializers

import (
	"log"

	"github.com/streadway/amqp"
)

func ConnectRabbitMQ(config *Config) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(config.RabbitMQUri)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	// Perform any additional setup or configurations

	return conn, ch
}
