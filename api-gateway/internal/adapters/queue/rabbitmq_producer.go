package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// RabbitMQProducer is a self-healing RabbitMQ producer.
// It owns the full connection lifecycle and automatically reconnects
// when the broker drops the connection (e.g. restart, network blip).
type RabbitMQProducer struct {
	amqpURL     string
	conn        *amqp.Connection
	channel     *amqp.Channel
	isConnected bool
	mu          sync.RWMutex
	done        chan struct{}
}

// NewRabbitMQProducer creates a new self-healing producer.
// It performs the initial connection synchronously and starts a background
// goroutine to watch for connection drops and auto-reconnect.
func NewRabbitMQProducer(amqpURL string) *RabbitMQProducer {
	p := &RabbitMQProducer{
		amqpURL: amqpURL,
		done:    make(chan struct{}),
	}

	// Initial connection (blocking — fail fast at startup if broker is down)
	if err := p.connect(); err != nil {
		log.Printf("⚠️ RabbitMQ initial connection failed: %v (will retry in background)", err)
	}

	// Start background reconnect watcher
	go p.watchConnection()

	return p
}

// connect dials RabbitMQ, opens a channel, and declares the exchange/queue/binding.
// Must be called while NOT holding the lock (it acquires it internally).
func (p *RabbitMQProducer) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close stale resources if any
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil && !p.conn.IsClosed() {
		_ = p.conn.Close()
	}

	conn, err := amqp.Dial(p.amqpURL)
	if err != nil {
		p.isConnected = false
		return fmt.Errorf("amqp.Dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		p.isConnected = false
		return fmt.Errorf("conn.Channel: %w", err)
	}

	// Declare exchange, queue, and binding so they exist after a broker restart
	if err := ch.ExchangeDeclare("doc_exchange", "direct", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		p.isConnected = false
		return fmt.Errorf("ExchangeDeclare: %w", err)
	}
	if _, err := ch.QueueDeclare("ocr_queue", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		p.isConnected = false
		return fmt.Errorf("QueueDeclare: %w", err)
	}
	if err := ch.QueueBind("ocr_queue", "ocr_queue", "doc_exchange", false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		p.isConnected = false
		return fmt.Errorf("QueueBind: %w", err)
	}

	p.conn = conn
	p.channel = ch
	p.isConnected = true
	log.Println("✅ RabbitMQ: Connected & exchange/queue declared")
	return nil
}

// watchConnection monitors the AMQP connection for drops and reconnects
// using exponential backoff (1s -> 2s -> 4s -> ... -> 30s max).
func (p *RabbitMQProducer) watchConnection() {
	for {
		// Wait until we have a live connection to watch
		p.mu.RLock()
		conn := p.conn
		connected := p.isConnected
		p.mu.RUnlock()

		if !connected || conn == nil {
			// Not connected yet — try to reconnect with backoff
			p.reconnectWithBackoff()
			continue
		}

		// Block until connection drops or is closed gracefully
		notifyClose := conn.NotifyClose(make(chan *amqp.Error, 1))

		select {
		case amqpErr, ok := <-notifyClose:
			if !ok {
				// Channel closed cleanly (e.g. Close() called)
				return
			}
			log.Printf("⚠️ RabbitMQ connection lost: %v — reconnecting...", amqpErr)
			p.mu.Lock()
			p.isConnected = false
			p.mu.Unlock()
			p.reconnectWithBackoff()

		case <-p.done:
			return
		}
	}
}

// reconnectWithBackoff retries connect() with exponential backoff.
func (p *RabbitMQProducer) reconnectWithBackoff() {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-p.done:
			return
		default:
		}

		log.Printf("🔄 RabbitMQ reconnecting in %v...", backoff)
		time.Sleep(backoff)

		if err := p.connect(); err != nil {
			log.Printf("❌ RabbitMQ reconnect failed: %v", err)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}
		return // Successfully reconnected
	}
}

// PublishJob publishes a job message to the ocr_queue.
// FAIL-FAST: Returns an error immediately if the connection is down,
// so the API Gateway can respond with HTTP 503 without blocking.
func (p *RabbitMQProducer) PublishJob(ctx context.Context, jobID string, docID string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Fail-fast check
	if !p.isConnected || p.channel == nil {
		return fmt.Errorf("rabbitmq: connection is not available (broker may be restarting)")
	}

	body, _ := json.Marshal(map[string]string{
		"job_id": jobID,
		"doc_id": docID,
	})

	// Inject current trace context into message headers
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	headers := amqp.Table{}
	for k, v := range carrier {
		headers[k] = v
	}

	err := p.channel.PublishWithContext(ctx,
		"doc_exchange", // Exchange
		"ocr_queue",    // Routing Key
		false,          // Mandatory
		false,          // Immediate
		amqp.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("❌ Failed to publish message: %v", err)
		return err
	}
	return nil
}

// Close gracefully shuts down the producer, stopping the watcher goroutine
// and closing the AMQP connection.
func (p *RabbitMQProducer) Close() {
	close(p.done)

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil && !p.conn.IsClosed() {
		_ = p.conn.Close()
	}
	p.isConnected = false
	log.Println("🔌 RabbitMQ producer closed")
}