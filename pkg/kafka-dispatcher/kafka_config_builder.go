package kafkadispatcher

import (
	"crypto/tls"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConfigBuilder provides a fluent interface for building Kafka ReaderConfig.
type KafkaConfigBuilder struct {
	config kafka.ReaderConfig
}

// NewKafkaConfigBuilder creates a new KafkaConfigBuilder with default values.
func NewKafkaConfigBuilder() *KafkaConfigBuilder {
	return &KafkaConfigBuilder{
		config: kafka.ReaderConfig{
			QueueCapacity:          100,
			MinBytes:               1,
			MaxBytes:               1024 * 1024, // 1MB
			MaxWait:                10 * time.Second,
			ReadBatchTimeout:       10 * time.Second,
			ReadLagInterval:        30 * time.Second,
			HeartbeatInterval:      3 * time.Second,
			CommitInterval:         0, // Manual commit by default
			PartitionWatchInterval: 5 * time.Second,
			SessionTimeout:         30 * time.Second,
			RebalanceTimeout:       30 * time.Second,
			JoinGroupBackoff:       5 * time.Second,
			RetentionTime:          -1, // Use broker default
			StartOffset:            kafka.FirstOffset,
			ReadBackoffMin:         100 * time.Millisecond,
			ReadBackoffMax:         1 * time.Second,
			MaxAttempts:            3,
		},
	}
}

// Brokers sets the list of broker addresses.
func (b *KafkaConfigBuilder) Brokers(brokers []string) *KafkaConfigBuilder {
	b.config.Brokers = brokers
	return b
}

// GroupID sets the consumer group ID.
func (b *KafkaConfigBuilder) GroupID(groupID string) *KafkaConfigBuilder {
	b.config.GroupID = groupID
	return b
}

// GroupTopics sets multiple topics for consumer group.
func (b *KafkaConfigBuilder) GroupTopics(topics []string) *KafkaConfigBuilder {
	b.config.GroupTopics = topics
	return b
}

// Topic sets the topic to read messages from.
func (b *KafkaConfigBuilder) Topic(topic string) *KafkaConfigBuilder {
	b.config.Topic = topic
	return b
}

// Partition sets the partition to read messages from.
func (b *KafkaConfigBuilder) Partition(partition int) *KafkaConfigBuilder {
	b.config.Partition = partition
	return b
}

// Dialer sets a custom dialer for connections.
func (b *KafkaConfigBuilder) Dialer(dialer *kafka.Dialer) *KafkaConfigBuilder {
	b.config.Dialer = dialer
	return b
}

// DefaultDialer sets a default dialer with timeout and TLS.
func (b *KafkaConfigBuilder) DefaultDialer(timeout time.Duration, tlsConfig *tls.Config) *KafkaConfigBuilder {
	b.config.Dialer = &kafka.Dialer{
		Timeout: timeout,
		TLS:     tlsConfig,
	}
	return b
}

// QueueCapacity sets the capacity of the internal message queue.
func (b *KafkaConfigBuilder) QueueCapacity(capacity int) *KafkaConfigBuilder {
	b.config.QueueCapacity = capacity
	return b
}

// MinBytes sets the minimum batch size that the consumer will accept.
func (b *KafkaConfigBuilder) MinBytes(minBytes int) *KafkaConfigBuilder {
	b.config.MinBytes = minBytes
	return b
}

// MaxBytes sets the maximum batch size that the consumer will accept.
func (b *KafkaConfigBuilder) MaxBytes(maxBytes int) *KafkaConfigBuilder {
	b.config.MaxBytes = maxBytes
	return b
}

// MaxWait sets the maximum amount of time to wait for new data.
func (b *KafkaConfigBuilder) MaxWait(maxWait time.Duration) *KafkaConfigBuilder {
	b.config.MaxWait = maxWait
	return b
}

// ReadBatchTimeout sets the amount of time to wait to fetch message from batch.
func (b *KafkaConfigBuilder) ReadBatchTimeout(timeout time.Duration) *KafkaConfigBuilder {
	b.config.ReadBatchTimeout = timeout
	return b
}

// ReadLagInterval sets the frequency at which the reader lag is updated.
func (b *KafkaConfigBuilder) ReadLagInterval(interval time.Duration) *KafkaConfigBuilder {
	b.config.ReadLagInterval = interval
	return b
}

// GroupBalancers sets the priority-ordered list of client-side consumer group balancing strategies.
func (b *KafkaConfigBuilder) GroupBalancers(balancers []kafka.GroupBalancer) *KafkaConfigBuilder {
	b.config.GroupBalancers = balancers
	return b
}

// HeartbeatInterval sets the frequency at which the reader sends heartbeat updates.
func (b *KafkaConfigBuilder) HeartbeatInterval(interval time.Duration) *KafkaConfigBuilder {
	b.config.HeartbeatInterval = interval
	return b
}

// CommitInterval sets the interval at which offsets are committed to the broker.
func (b *KafkaConfigBuilder) CommitInterval(interval time.Duration) *KafkaConfigBuilder {
	b.config.CommitInterval = interval
	return b
}

// AutoCommit enables automatic offset commits with the specified interval.
func (b *KafkaConfigBuilder) AutoCommit(interval time.Duration) *KafkaConfigBuilder {
	b.config.CommitInterval = interval
	return b
}

// ManualCommit disables automatic offset commits (manual commit only).
func (b *KafkaConfigBuilder) ManualCommit() *KafkaConfigBuilder {
	b.config.CommitInterval = 0
	return b
}

// PartitionWatchInterval sets how often a reader checks for partition changes.
func (b *KafkaConfigBuilder) PartitionWatchInterval(interval time.Duration) *KafkaConfigBuilder {
	b.config.PartitionWatchInterval = interval
	return b
}

// WatchPartitionChanges enables polling for partition changes and rebalancing.
func (b *KafkaConfigBuilder) WatchPartitionChanges(watch bool) *KafkaConfigBuilder {
	b.config.WatchPartitionChanges = watch
	return b
}

// SessionTimeout sets the length of time before the coordinator considers the consumer dead.
func (b *KafkaConfigBuilder) SessionTimeout(timeout time.Duration) *KafkaConfigBuilder {
	b.config.SessionTimeout = timeout
	return b
}

// RebalanceTimeout sets the length of time the coordinator will wait for members to join.
func (b *KafkaConfigBuilder) RebalanceTimeout(timeout time.Duration) *KafkaConfigBuilder {
	b.config.RebalanceTimeout = timeout
	return b
}

// JoinGroupBackoff sets the length of time to wait between re-joining the consumer group.
func (b *KafkaConfigBuilder) JoinGroupBackoff(backoff time.Duration) *KafkaConfigBuilder {
	b.config.JoinGroupBackoff = backoff
	return b
}

// RetentionTime sets the length of time the consumer group will be saved by the broker.
func (b *KafkaConfigBuilder) RetentionTime(retention time.Duration) *KafkaConfigBuilder {
	b.config.RetentionTime = retention
	return b
}

// StartOffset sets the offset from which to start consuming when no committed offset exists.
func (b *KafkaConfigBuilder) StartOffset(offset int64) *KafkaConfigBuilder {
	b.config.StartOffset = offset
	return b
}

// StartFromBeginning sets the start offset to the beginning.
func (b *KafkaConfigBuilder) StartFromBeginning() *KafkaConfigBuilder {
	b.config.StartOffset = kafka.FirstOffset
	return b
}

// StartFromEnd sets the start offset to the end.
func (b *KafkaConfigBuilder) StartFromEnd() *KafkaConfigBuilder {
	b.config.StartOffset = kafka.LastOffset
	return b
}

// ReadBackoffMin sets the smallest amount of time the reader will wait before polling.
func (b *KafkaConfigBuilder) ReadBackoffMin(min time.Duration) *KafkaConfigBuilder {
	b.config.ReadBackoffMin = min
	return b
}

// ReadBackoffMax sets the maximum amount of time the reader will wait before polling.
func (b *KafkaConfigBuilder) ReadBackoffMax(max time.Duration) *KafkaConfigBuilder {
	b.config.ReadBackoffMax = max
	return b
}

// Logger sets the logger for internal changes.
func (b *KafkaConfigBuilder) Logger(logger kafka.Logger) *KafkaConfigBuilder {
	b.config.Logger = logger
	return b
}

// ErrorLogger sets the logger for errors.
func (b *KafkaConfigBuilder) ErrorLogger(logger kafka.Logger) *KafkaConfigBuilder {
	b.config.ErrorLogger = logger
	return b
}

// IsolationLevel sets the visibility of transactional records.
func (b *KafkaConfigBuilder) IsolationLevel(level kafka.IsolationLevel) *KafkaConfigBuilder {
	b.config.IsolationLevel = level
	return b
}

// MaxAttempts sets the limit of connection attempts.
func (b *KafkaConfigBuilder) MaxAttempts(attempts int) *KafkaConfigBuilder {
	b.config.MaxAttempts = attempts
	return b
}

// OffsetOutOfRangeError enables returning an error on OffsetOutOfRange instead of retrying.
func (b *KafkaConfigBuilder) OffsetOutOfRangeError(enable bool) *KafkaConfigBuilder {
	b.config.OffsetOutOfRangeError = enable
	return b
}

// Build returns the configured kafka.ReaderConfig.
func (b *KafkaConfigBuilder) Build() kafka.ReaderConfig {
	return b.config
}

// BuildReader creates a new kafka.Reader with the configured settings.
func (b *KafkaConfigBuilder) BuildReader() *kafka.Reader {
	return kafka.NewReader(b.config)
}
