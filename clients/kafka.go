package clients

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"vivek-ray/constants"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type KafkaConfig struct {
	Brokers       []string
	Username      string
	Password      string
	SASLMechanism string
	Version       string
}

type KafkaService struct {
	config       *KafkaConfig
	saramaConfig *sarama.Config
	client       sarama.Client
	producers    map[string]sarama.AsyncProducer
	mu           sync.RWMutex
}

type KafkaWriter struct {
	producer sarama.AsyncProducer
	topic    string
}
type KafkaReader struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	groupID       string
	handler       *consumerGroupHandler
}

func NewKafkaService(config *KafkaConfig) (*KafkaService, error) {
	saramaConfig := sarama.NewConfig()

	versionStr := config.Version
	version, err := sarama.ParseKafkaVersion(versionStr)
	if err != nil {
		return nil, constants.ErrorWrap(constants.ErrKafkaInvalidVersion, err)
	}
	saramaConfig.Version = version

	if config.Username != "" && config.Password != "" {
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.User = config.Username
		saramaConfig.Net.SASL.Password = config.Password
		saramaConfig.Net.SASL.Mechanism = sarama.SASLMechanism(config.SASLMechanism)
		saramaConfig.Net.TLS.Enable = true
	}

	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 3
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true

	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewClient(config.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &KafkaService{
		config:       config,
		saramaConfig: saramaConfig,
		client:       client,
		producers:    make(map[string]sarama.AsyncProducer),
	}, nil
}

func (k *KafkaService) InitWriter(topic string) (*KafkaWriter, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if producer, exists := k.producers[topic]; exists {
		return &KafkaWriter{producer: producer, topic: topic}, nil
	}

	producer, err := sarama.NewAsyncProducerFromClient(k.client)
	if err != nil {
		return nil, err
	}

	k.producers[topic] = producer
	return &KafkaWriter{producer: producer, topic: topic}, nil
}

func (w *KafkaWriter) BulkWrite(values any) error {
	rv := reflect.ValueOf(values)
	if rv.Kind() != reflect.Slice {
		return constants.ErrKafkaInvalidInput
	}

	for i := 0; i < rv.Len(); i++ {
		data, err := json.Marshal(rv.Index(i).Interface())
		if err != nil {
			return err
		}
		w.producer.Input() <- &sarama.ProducerMessage{
			Topic: w.topic,
			Key:   sarama.StringEncoder(uuid.NewString()),
			Value: sarama.ByteEncoder(data),
		}
	}
	return nil
}

func (w *KafkaWriter) Close() error {
	return w.producer.Close()
}

type consumerGroupHandler struct {
	handler func(message *sarama.ConsumerMessage) error
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.handler(msg); err != nil {
			log.Error().Err(err).
				Str("topic", msg.Topic).
				Int32("partition", msg.Partition).
				Int64("offset", msg.Offset).
				Msg("failed to process message")
			continue
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func (k *KafkaService) InitReader(topic, groupID string) (*KafkaReader, error) {
	consumerGroup, err := sarama.NewConsumerGroupFromClient(groupID, k.client)
	if err != nil {
		return nil, err
	}

	return &KafkaReader{
		consumerGroup: consumerGroup,
		topic:         topic,
		groupID:       groupID,
		handler:       &consumerGroupHandler{},
	}, nil
}

func (r *KafkaReader) Read(ctx context.Context, handler func(message *sarama.ConsumerMessage) error) error {
	r.handler.handler = handler

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := r.consumerGroup.Consume(ctx, []string{r.topic}, r.handler); err != nil {
				log.Error().Err(err).Msg("consumer group error")
				return err
			}
		}
	}
}

func (r *KafkaReader) Close() error {
	return r.consumerGroup.Close()
}

func (k *KafkaService) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	for _, producer := range k.producers {
		if err := producer.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close producer")
		}
	}

	return k.client.Close()
}

func (k *KafkaService) Client() sarama.Client {
	return k.client
}
