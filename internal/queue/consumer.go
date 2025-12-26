package queue

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"base-go-app/internal/broadcast"
	"base-go-app/internal/config"
	"base-go-app/internal/tasks"
	"base-go-app/internal/webhook"

	amqp "github.com/rabbitmq/amqp091-go"
)

// rabbitConnected indicates whether the consumer has an active RabbitMQ connection
var rabbitConnected int32 // 0 = false, 1 = true

func RabbitConnected() bool {
	return atomic.LoadInt32(&rabbitConnected) == 1
}

// SetRabbitConnectedForTests is a helper used by tests to set rabbit state.
func SetRabbitConnectedForTests(v bool) {
	if v {
		atomic.StoreInt32(&rabbitConnected, 1)
	} else {
		atomic.StoreInt32(&rabbitConnected, 0)
	}
}

// StartConsumer starts the consumer loop in a background goroutine and returns
// a channel that will be closed when the consumer exits (typically because ctx
// was canceled).
func StartConsumer(ctx context.Context, cfg *config.Config) <-chan struct{} {
	done := make(chan struct{})

	// Initialize dependencies
	broadcaster := broadcast.NewSockudoBroadcaster()
	webhookClient := webhook.NewOAuthClient(
		os.Getenv("WEBHOOK_OAUTH_TOKEN_URL"),
		os.Getenv("WEBHOOK_OAUTH_CLIENT_ID"),
		os.Getenv("WEBHOOK_OAUTH_CLIENT_SECRET"),
		os.Getenv("WEBHOOK_OAUTH_SCOPE"),
	)
	dispatcher := tasks.NewDispatcher(broadcaster, webhookClient)

	// Worker pool config
	concurrency := 10
	if s := os.Getenv("WORKER_CONCURRENCY"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			concurrency = v
		}
	}
	bufferSize := 100
	if s := os.Getenv("TASK_CHANNEL_BUFFER"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			bufferSize = v
		}
	}

	taskCh := make(chan amqp.Delivery, bufferSize)
	var wg sync.WaitGroup

	// Shared channel for publishing retries
	var (
		chMu      sync.RWMutex
		currentCh *amqp.Channel
	)

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case d, ok := <-taskCh:
					if !ok {
						return
					}
					// Process task
					res := dispatcher.Dispatch(ctx, d.Body)
					if res.Success {
						d.Ack(false)
					} else if res.Retry {
						// Attempt to republish with incremented attempt count
						var payload tasks.TaskPayload
						if err := json.Unmarshal(d.Body, &payload); err == nil {
							payload.Attempt = res.RetryAttempt
							newBody, _ := json.Marshal(payload)

							// Calculate backoff (exponential: 2^(attempt-1) seconds)
							// e.g., attempt 1 (retry 1) -> 1s, retry 2 -> 2s, retry 3 -> 4s
							backoffMs := int64(1000 * (1 << (payload.Attempt - 1)))

							chMu.RLock()
							pubCh := currentCh
							chMu.RUnlock()

							if pubCh != nil {
								err := pubCh.Publish(
									d.Exchange,
									d.RoutingKey,
									false, // mandatory
									false, // immediate
									amqp.Publishing{
										ContentType: "application/json",
										Body:        newBody,
										Headers: amqp.Table{
											"x-delay": backoffMs, // For rabbitmq_delayed_message_exchange
										},
									},
								)
								if err == nil {
									d.Ack(false)
									continue
								}
								log.Printf("Failed to republish retry: %v", err)
							}
						}
						// Fallback: Nack without requeue (DLQ)
						d.Nack(false, false)
					} else {
						// Fatal error
						d.Nack(false, false)
					}
				}
			}
		}(i)
	}

	go func() {
		defer close(done)
		defer wg.Wait() // Wait for workers to finish
		defer close(taskCh)

		delay := 2 * time.Second
		for {
			select {
			case <-ctx.Done():
				log.Println("StartConsumer: context canceled, shutting down consumer")
				atomic.StoreInt32(&rabbitConnected, 0)
				return
			default:
			}

			log.Printf("Attempting RabbitMQ connect...")
			conn, err := amqp.Dial(cfg.GetRabbitMQURL())
			if err != nil {
				log.Printf("RabbitMQ connect failed: %v", err)
				// backoff
				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
				}
				if delay < 30*time.Second {
					delay *= 2
					if delay > 30*time.Second {
						delay = 30 * time.Second
					}
				}
				continue
			}

			// Connected
			atomic.StoreInt32(&rabbitConnected, 1)
			log.Println("Connected to RabbitMQ")

			ch, err := conn.Channel()
			if err != nil {
				log.Printf("Failed to open a channel: %v", err)
				_ = conn.Close()
				atomic.StoreInt32(&rabbitConnected, 0)
				continue
			}

			// Update shared channel
			chMu.Lock()
			currentCh = ch
			chMu.Unlock()

			// Set QoS
			if err := ch.Qos(concurrency*2, 0, false); err != nil {
				log.Printf("Failed to set QoS: %v", err)
			}

			queueName := "logger"
			exchangeName := "celery" // Keeping legacy name for now, or switch to "tasks"
			routingKey := "logger"   // or task.log_db

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
				log.Printf("Failed to declare exchange: %v", err)
				ch.Close()
				_ = conn.Close()
				atomic.StoreInt32(&rabbitConnected, 0)
				continue
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
				log.Printf("Failed to declare a queue: %v", err)
				ch.Close()
				_ = conn.Close()
				atomic.StoreInt32(&rabbitConnected, 0)
				continue
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
				log.Printf("Failed to bind queue: %v", err)
				ch.Close()
				_ = conn.Close()
				atomic.StoreInt32(&rabbitConnected, 0)
				continue
			}

			msgs, err := ch.Consume(
				q.Name, // queue
				"",     // consumer
				false,  // auto-ack (FALSE now, manual ack in worker)
				false,  // exclusive
				false,  // no-local
				false,  // no-wait
				nil,    // args
			)
			if err != nil {
				log.Printf("Failed to register a consumer: %v", err)
				ch.Close()
				_ = conn.Close()
				atomic.StoreInt32(&rabbitConnected, 0)
				continue
			}

			// Reset delay after successful connection
			delay = 2 * time.Second

			// Process messages; when msgs channel closes we attempt to reconnect
			notifyClose := conn.NotifyClose(make(chan *amqp.Error))
			for {
				select {
				case <-ctx.Done():
					log.Println("Context canceled while consuming, closing consumer")
					chMu.Lock()
					currentCh = nil
					chMu.Unlock()
					ch.Close()
					_ = conn.Close()
					atomic.StoreInt32(&rabbitConnected, 0)
					return
				case err := <-notifyClose:
					log.Printf("RabbitMQ connection closed: %v", err)
					goto Reconnect
				case d, ok := <-msgs:
					if !ok {
						// Channel closed
						log.Println("msgs channel closed")
						goto Reconnect
					}
					// Push to worker pool
					select {
					case taskCh <- d:
					case <-ctx.Done():
						return
					}
				}
			}
		Reconnect:
			// msgs channel closed or connection lost
			log.Println("RabbitMQ consumer disconnected, will attempt reconnect")
			chMu.Lock()
			currentCh = nil
			chMu.Unlock()
			ch.Close()
			_ = conn.Close()
			atomic.StoreInt32(&rabbitConnected, 0)
			// loop and retry
		}
	}()
	return done
}
