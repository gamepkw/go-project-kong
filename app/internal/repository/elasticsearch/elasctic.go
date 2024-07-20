package elastic

import (
	"github.com/elastic/go-elasticsearch/v7"
)

type elastic struct {
	es *elasticsearch.Client
}

func New(
	es *elasticsearch.Client,
) Elastic {
	return &elastic{
		es: es,
	}
}

type Elastic interface {
}
