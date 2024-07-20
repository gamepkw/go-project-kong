package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"main/app/pkg/kafka/consumer"
	"sync"

	"github.com/IBM/sarama"
)

func (s *service) ConsumeMessage(ctx context.Context) error {

	topics := []string{"test_topic"}
	consumer := consumer.Consumer{
		Messages: make(chan *sarama.ConsumerMessage),
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := s.consumer.Consume(ctx, topics, &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	for msg := range consumer.Messages {
		fmt.Println("Received message:", string(msg.Value))
	}

	return nil
}
