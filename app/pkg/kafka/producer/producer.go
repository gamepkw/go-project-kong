package producer

// SIGUSR1 toggle the pause/resume consumption
import (
	"fmt"
	"log"
	"main/app/config"
	"time"

	_ "net/http/pprof"

	"github.com/IBM/sarama"
)

func New(conf *config.Config) *sarama.AsyncProducer {
	log.Println("Starting a new Sarama producer")

	version, err := sarama.ParseKafkaVersion(conf.Secrets.KafkaVersion)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.Version = version
	config.Producer.Idempotent = true
	config.Producer.Return.Errors = false
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Transaction.Retry.Backoff = 10
	config.Producer.Transaction.ID = "txn_producer"
	config.Producer.Timeout = 30 * time.Second
	config.Net.MaxOpenRequests = 1

	producer, err := sarama.NewAsyncProducer([]string{conf.Secrets.KafkaBroker}, config)
	if err != nil {
		panic(fmt.Sprintf("failed to create producer: %v", err))
	}

	producer.Input()

	return &producer
}

func ProduceMessage(producer sarama.AsyncProducer, topics []string, message string) error {

	err := producer.BeginTxn()
	if err != nil {
		log.Printf("unable to start txn %s\n", err)
		return err
	}

	var recordsNumber int64 = 10
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
