package kafkadispatcher

import (
	"context"
	"encoding/json"
	"log"
	"runtime"
	"sync"

	"github.com/segmentio/kafka-go"
)

// Dispatcher manages handlers and feeds Kafka events to them.
type Dispatcher struct {
	handlers   []Handler
	reader     *kafka.Reader
	workerPool chan struct{}
	wg         sync.WaitGroup
}

// NewDispatcher creates a dispatcher with Kafka config.
func NewDispatcher(brokers []string, topic, groupID string, handlers ...Handler) *Dispatcher {
	return &Dispatcher{
		handlers: handlers,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		}),
		workerPool: make(chan struct{}, runtime.NumCPU()),
	}
}

// RegisterHandler adds a handler to the dispatcher.
func (d *Dispatcher) RegisterHandler(h Handler) {
	d.handlers = append(d.handlers, h)
}

// Start begins consuming messages and dispatching to handlers concurrently.
func (d *Dispatcher) Start(ctx context.Context) error {
	for {
		m, err := d.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		var event Event
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Println("failed to unmarshal event:", err)
			continue
		}

		// acquire a worker slot
		d.workerPool <- struct{}{}
		d.wg.Add(1)

		go func(e Event) {
			defer func() {
				<-d.workerPool
				d.wg.Done()
			}()

			for _, h := range d.handlers {
				if err := h.Handle(ctx, e); err != nil {
					log.Println("handler error:", err)
				}
			}
		}(event)
	}
}

// Close gracefully shuts down the dispatcher.
func (d *Dispatcher) Close() error {
	if err := d.reader.Close(); err != nil {
		return err
	}
	d.wg.Wait()
	return nil
}
