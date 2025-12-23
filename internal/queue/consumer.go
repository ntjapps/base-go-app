package queue

import (
	"base-go-app/internal/config"
	"base-go-app/internal/tasks"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartConsumer(cfg *config.Config) {
	conn, err := amqp.Dial(cfg.GetRabbitMQURL())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	queueName := "logger"
	exchangeName := "celery"
	routingKey := "logger"

	// Declare Exchange
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	// Declare Queue
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Bind Queue
	err = ch.QueueBind(
		q.Name,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	forever := make(chan struct{})

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			// We assume all messages in this queue are for the logger task
			err := tasks.HandleLoggerTask(d.Body)
			if err != nil {
				log.Printf("Error handling task: %v", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages in queue %s. To exit press CTRL+C", queueName)
	<-forever
}
