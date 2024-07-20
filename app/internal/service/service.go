package service

import (
	"context"
	"main/app/internal/repository/cache"
	"main/app/internal/repository/database"
	elastic "main/app/internal/repository/elasticsearch"

	"github.com/IBM/sarama"
)

type service struct {
	db       database.Database
	cache    cache.Cache
	es       elastic.Elastic
	consumer sarama.ConsumerGroup
	producer sarama.AsyncProducer
}

func New(
	db database.Database,
	cache cache.Cache,
	es elastic.Elastic,
	consumer sarama.ConsumerGroup,
	producer sarama.AsyncProducer) Service {
	return &service{
		db:       db,
		cache:    cache,
		es:       es,
		consumer: consumer,
		producer: producer,
	}
}

type Service interface {
	ProduceMessage(ctx context.Context) error
	ConsumeMessage(ctx context.Context) error
}
