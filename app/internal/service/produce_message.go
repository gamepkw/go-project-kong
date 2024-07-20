package service

import (
	"context"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

func (s *service) ProduceMessage(ctx context.Context) error {

	var producers = 1
	topics := []string{"test_topic"}
	message := "hello world"

	var wg sync.WaitGroup
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					produceMessage(s.producer, topics, message)
				}
			}
		}()
	}
	return nil
}

func produceMessage(producer sarama.AsyncProducer, topics []string, message string) error {

	err := producer.BeginTxn()
	if err != nil {
		log.Printf("unable to start txn %s\n", err)
		return err
	}

	var recordsNumber int64 = 1
	var i int64
	for i = 0; i < recordsNumber; i++ {
		producer.Input() <- &sarama.ProducerMessage{Topic: topics[0], Key: nil, Value: sarama.StringEncoder(message)}
	}

	err = producer.CommitTxn()
	if err != nil {
		log.Printf("Producer: unable to commit txn %s\n", err)
		for {
			if producer.TxnStatus()&sarama.ProducerTxnFlagFatalError != 0 {
				log.Printf("Producer: producer is in a fatal state, need to recreate it")
				break
			}
			if producer.TxnStatus()&sarama.ProducerTxnFlagAbortableError != 0 {
				err = producer.AbortTxn()
				if err != nil {
					log.Printf("Producer: unable to abort transaction: %+v", err)
					continue
				}
				break
			}

			err = producer.CommitTxn()
			if err != nil {
				log.Printf("Producer: unable to commit txn %s\n", err)
				continue
			}
		}
		return err
	}
	return nil
}
