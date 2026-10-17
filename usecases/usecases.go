package usecases

import (
	"github.com/RandySteven/go-kopi/caches"
	nsq_client "github.com/RandySteven/go-cook/nsq"
	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/go-kopi/repositories"
)

type Usecases struct {
	UserUsecase UserUsecase
}

func NewUsecases(repositories *repositories.Repositories,
	redis *caches.Caches,
	nsq nsq_client.Nsq,
	temporal temporal_client.Temporal,
) *Usecases {
	return &Usecases{}
}
