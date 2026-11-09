package consumers

import (
	"context"
	"fmt"
	"log"

	nsq_client "github.com/RandySteven/go-cook/nsq"
	"github.com/nsqio/go-nsq"
)

type (
	Runners struct {
		channel      string
		nsq          nsq_client.Nsq
		Consumers    []Consumer
		RunConsumers RunConsumer
	}

	ConsumerFunc func(ctx context.Context, nsqMessage *nsq.Message) error
)

func InitRunner(nsq nsq_client.Nsq, channel string) *Runners {
	return &Runners{
		channel:      channel,
		nsq:          nsq,
		RunConsumers: make(map[string]ConsumerFunc),
	}
}

func (r *Runners) Run(ctx context.Context) error {
	errChan := make(chan error, len(r.RunConsumers))

	for topic, consumer := range r.RunConsumers {
		go func(topic string, consumer ConsumerFunc) {
			log.Println(`execute consumer `, consumer)
			err := r.nsq.RegisterConsumer(topic, r.channel, func(msgCtx context.Context, nsqMessage *nsq.Message) error {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Recovered from panic in consumer %s: %v", topic, r)
					}
				}()

				if err := consumer(msgCtx, nsqMessage); err != nil {
					log.Printf("Error in consumer %s: %v", topic, err)
					return err
				}

				return nil
			})
			if err != nil {
				errChan <- fmt.Errorf("failed to register consumer for topic %s: %w", topic, err)
			}
		}(topic, consumer)
	}

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
