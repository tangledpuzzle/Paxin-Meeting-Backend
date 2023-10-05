package utils

import (
	"log"

	"github.com/streadway/amqp"
)


func PublishMessage(ch *amqp.Channel, queueName string, message string) error {
	// Example code for publishing a message
	err := ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return err
	}

	return nil
}
func ConsumeMessages(ch *amqp.Channel, conn *amqp.Connection, queueName string) error {
	// Example code for consuming messages
	msgs, err := ch.Consume(
		queueName,
		"",
		false, // Auto-acknowledge disabled
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Create a map to store delivery tags
	deliveryTags := make(map[uint64]bool)

	// Consume messages in a loop until the queue is empty
	for msg := range msgs {
		// Обработка полученного сообщения
		// emailContent := string(msg.Body)
		// SendEmailDelayedSuccess(emailContent)

		// Store the delivery tag
		deliveryTag := msg.DeliveryTag
		deliveryTags[deliveryTag] = true

		// Дополнительные действия после обработки сообщения
		// ...

		// Acknowledge the message
		if err := ch.Ack(deliveryTag, false); err != nil {
			log.Printf("Failed to acknowledge message: %s", err)
			// Handle the error if needed
		}

		// Check if there are more messages in the queue
		queueEmpty, err := isQueueEmpty(ch, queueName)
		if err != nil {
			log.Printf("Failed to check queue status: %s", err)
			// Handle the error if needed
		}

		if queueEmpty {
			break // Exit the loop if the queue is empty
		}
	}

	// Purge the queue after all messages have been processed
	numPurged, err := ch.QueuePurge(queueName, false)
	if err != nil {
		log.Printf("Failed to purge queue: %s", err)
	} else {
		log.Printf("Purged %d messages from the queue '%s'", numPurged, queueName)
	}

	// Close the channel outside the loop
	if err := ch.Close(); err != nil {
		log.Printf("Failed to close channel: %s", err)
	}

	// Close the connection outside the loop
	if err := conn.Close(); err != nil {
		log.Printf("Failed to close connection: %s", err)
	}

	return nil
}
// Function to check if the queue is empty
func isQueueEmpty(ch *amqp.Channel, queueName string) (bool, error) {
	queue, err := ch.QueueInspect(queueName)
	if err != nil {
		return false, err
	}
	return queue.Messages == 0, nil
}
// func SendEmailDelayedSuccess(emailContent string) error {
// 	// Simulate sending email with a delay
// 	time.Sleep(5 * time.Second)

// 	// Simulate failures
// 	for i := 0; i < 10; i++ {
// 		log.Printf("Failed to send email attempt #%d", i+1)
// 		time.Sleep(1 * time.Second)
// 	}

// 	// Simulate successful sending
// 	log.Println("Successfully sent email")
// 	return nil
// }

