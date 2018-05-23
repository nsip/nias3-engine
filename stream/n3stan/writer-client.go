// writer-client.go

package n3stan

import (
	"log"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
)

// multiplex the connection for multiple subscriptions
var writerConn stan.Conn

type writerClient struct {
	StanConfig                     // config for stan server access
	SubscriptionConfig             // config for topic subscription
	StanConn           stan.Conn   // connection to stan
	MsgC               chan []byte // channel on which messages for the stream can be published
	ErrC               chan error  // channel for errors encountered during publishing
}

func NewWriterClient(config StanConfig) (*writerClient, error) {

	wc := &writerClient{StanConfig: config,
		MsgC: make(chan []byte),
		ErrC: make(chan error, 1),
	}
	err := wc.connectToStan()
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to streaming server")
	}

	return wc, nil
}

func (wc *writerClient) connectToStan() error {

	if writerConn == nil {

		natsOpts := nats.GetDefaultOptions()
		natsOpts.AsyncErrorCB = func(c *nats.Conn, s *nats.Subscription, err error) {
			log.Println("got stan nats async error:", err)
		}

		natsOpts.DisconnectedCB = func(c *nats.Conn) {
			log.Println("stan Nats disconnected")
		}

		natsOpts.ReconnectedCB = func(c *nats.Conn) {
			log.Println("stan Nats reconnected")
		}
		natsOpts.Url = wc.Url

		natsCon, err := natsOpts.Connect()
		if err != nil {
			return errors.Wrap(err, "nats connect failed")
		}

		var stanConn stan.Conn
		if stanConn, err = stan.Connect(
			wc.ClusterId,
			wc.ClientId,
			stan.NatsConn(natsCon),
		); err != nil {
			return errors.Wrap(err, "stan connect failed")
		}
		writerConn = stanConn
	}

	wc.StanConn = writerConn

	return nil
}

func (wc *writerClient) subscribe(config SubscriptionConfig) error {

	wc.SubscriptionConfig = config

	// start listening for messages
	go wc.listen()

	return nil
}

func (wc *writerClient) listen() {
	for msg := range wc.MsgC {
		// log.Println("writer client listener received message")
		wc.publish(msg)
	}
}

func (wc *writerClient) publish(msg []byte) {

	_, err := wc.StanConn.PublishAsync(wc.SubscriptionConfig.Topic, msg, wc.ackHandler())
	if err != nil {
		wc.ErrC <- err
	}
	// log.Println("writer client published message")
}

func (wc *writerClient) ackHandler() func(nuid string, err error) {
	return func(ackedNuid string, err error) {
		if err != nil {
			ackErr := errors.Wrap(err, "error publishing message")
			wc.ErrC <- ackErr
			// log.Printf("Warning: error publishing msg id %s: %v\n", ackedNuid, err.Error())
		} else {
			// log.Printf("Received ack for msg id %s\n", ackedNuid)
		}
	}
}
