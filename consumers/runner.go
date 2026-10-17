package consumers

import (
	"context"
	"fmt"
	"log"

	nsq_client "github.com/RandySteven/go-cook/nsq"
)

type (
	Runners struct {
		channel      string
		nsq          nsq_client.Nsq
		ConsumerFunc []ConsumerFunc
		RunConsumers RunConsumer
	}

	ConsumerFunc func(ctx context.Context, msgBody string) error
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
			err := r.nsq.RegisterConsumer(topic, r.channel, func(msgCtx context.Context, body string) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Recovered from panic in consumer %s: %v", topic, r)
					}
				}()

				if err := consumer(msgCtx, body); err != nil {
					log.Printf("Error in consumer %s: %v", topic, err)
				}
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
