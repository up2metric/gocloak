package kafkadispatcher

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	// Singleton Pattern for the dispatcher
	// instance is the singleton instance of the dispatcher.
	instance *Dispatcher
	// Once is used to ensure that the dispatcher is created only once.
	once sync.Once
)

// Dispatcher manages handlers and feeds Kafka events to them.
type Dispatcher struct {
	handlers []Handler
	reader   *kafka.Reader

	mu     sync.Mutex
	cancel context.CancelFunc
}

// NewDispatcher creates a new dispatcher with the given builder and handlers.
func NewDispatcher(builder *KafkaConfigBuilder, handlers ...Handler) *Dispatcher {
	once.Do(func() {
		builder = builder.
			ManualCommit().
			DefaultDialer(10*time.Second, &tls.Config{})

		instance = &Dispatcher{
			handlers: handlers,
			reader:   builder.BuildReader(),
		}
	})

	return instance
}

// RegisterHandler registers a new handler.
// Mutex is used to synchronize access to the handlers slice.
func (d *Dispatcher) RegisterHandler(h Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers = append(d.handlers, h)
}

// Start begins consuming messages sequentially and in order.
func (d *Dispatcher) Start(ctx context.Context) {
	var internalCtx context.Context
	internalCtx, d.cancel = context.WithCancel(ctx)

	log.Println("Dispatcher started in sequential mode...")
	for {
		m, err := d.reader.FetchMessage(internalCtx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("Context cancelled. Shutting down consumer loop.")
				break
			}
			log.Println("failed to fetch message:", err)
			continue
		}

		var event Event
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Failed to unmarshal event (offset %d): %v. Skipping poison pill.", m.Offset, err)
			if err := d.reader.CommitMessages(internalCtx, m); err != nil {
				log.Println("Failed to commit poison pill message:", err)
			}
			continue
		}

		var handlerFailed bool = false
		d.mu.Lock()
		for _, h := range d.handlers {
			if err := h.Handle(internalCtx, event); err != nil {
				log.Printf("Handler error for event (offset %d): %v. Message will NOT be committed.", m.Offset, err)
				handlerFailed = true
				break
			}
		}
		d.mu.Unlock()

		if !handlerFailed {
			if err := d.reader.CommitMessages(internalCtx, m); err != nil {
				log.Println("failed to commit message:", err)
			}
		}
	}
	log.Println("Dispatcher consumer loop finished.")
}

// Close closes the dispatcher.
func (d *Dispatcher) Close() error {
	log.Println("Closing dispatcher...")
	if d.cancel != nil {
		d.cancel()
	}
	return d.reader.Close()
}
