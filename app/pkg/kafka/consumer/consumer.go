package consumer

// SIGUSR1 toggle the pause/resume consumption
import (
	"log"
	"main/app/config"
	"strings"

	"github.com/IBM/sarama"
)

func New(conf *config.Config) *sarama.ConsumerGroup {
	log.Println("Starting a new Sarama consumer")

	version, err := sarama.ParseKafkaVersion(conf.Secrets.KafkaVersion)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.Version = version

	oldest := true
	if oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	consumerGroup, err := sarama.NewConsumerGroup(strings.Split(conf.Secrets.KafkaBroker, ","), conf.Secrets.KafkaConsumerGroup, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	return &consumerGroup
}

type Consumer struct {
	Messages chan *sarama.ConsumerMessage
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// close(consumer.messages)
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}
			// log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			consumer.Messages <- message
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

// func ConsumeMessage(ctx context.Context, cg sarama.ConsumerGroup, topics []string) (<-chan *sarama.ConsumerMessage, error) {

// 	consumer := Consumer{
// 		Messages: make(chan *sarama.ConsumerMessage),
// 	}

// 	wg := &sync.WaitGroup{}
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		for {
// 			if err := cg.Consume(ctx, strings.Split(topics[0], ","), &consumer); err != nil {
// 				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
// 					return
// 				}
// 				log.Panicf("Error from consumer: %v", err)
// 			}
// 			if ctx.Err() != nil {
// 				return
// 			}
// 		}
// 	}()
// 	return consumer.Messages, nil
// }
