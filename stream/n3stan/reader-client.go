// reader-client.go

package n3stan

import (
	"log"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
)

// multiplex the connection for multiple subscriptions
var readerConn stan.Conn

type readerClient struct {
	StanConfig                           // config for stan server access
	SubscriptionConfig                   // config for topic subscription
	StanConn           stan.Conn         // connection to stan
	MsgC               chan *stan.Msg    // channel on which messages from the stream can be read
	ErrC               chan error        // channel for errors encountered during reads
	sub                stan.Subscription // subscription used to attach to topic
}

func NewReaderClient(config StanConfig) (*readerClient, error) {

	rc := &readerClient{StanConfig: config,
		MsgC: make(chan *stan.Msg),
		ErrC: make(chan error, 1),
	}
	err := rc.connectToStan()
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to streaming server")
	}

	return rc, nil
}

func (rc *readerClient) connectToStan() error {

	if readerConn == nil {

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
		natsOpts.Url = rc.Url

		natsCon, err := natsOpts.Connect()
		if err != nil {
			return errors.Wrap(err, "nats connect failed")
		}

		var stanConn stan.Conn
		if stanConn, err = stan.Connect(
			rc.ClusterId,
			rc.ClientId,
			stan.NatsConn(natsCon),
		); err != nil {
			return errors.Wrap(err, "stan connect failed")
		}
		readerConn = stanConn
	}

	rc.StanConn = readerConn

	return nil
}

func (rc *readerClient) subscribe(config SubscriptionConfig) error {

	rc.SubscriptionConfig = config

	var err error

	rc.sub, err = rc.StanConn.Subscribe(
		rc.Topic,
		rc.handlerFunc,
		stan.DurableName(rc.DurableName),
		stan.SetManualAckMode(),
	)

	if err != nil {
		return err
	}

	return nil
}

func (rc *readerClient) nondurableSubscribe(config SubscriptionConfig) error {

	rc.SubscriptionConfig = config

	var err error

	rc.sub, err = rc.StanConn.Subscribe(
		rc.Topic,
		rc.handlerFunc,
		stan.DeliverAllAvailable())

	if err != nil {
		return err
	}

	return nil

}

//
// handles inbouund messages, simply pushes to C (channel) for onward consumption
//
// NOTE:
// manual Ack mode is set for the subscription,
// msg will not be acked here, but by downstream processors
// that need to ack/noAck based on business logic requirements.
// Stan msgs retain a backlink to their subscription.
//
func (rc *readerClient) handlerFunc(msg *stan.Msg) {

	// TODO: maybe make this goroutine for faster asynch.
	// log.Println("\treader client got a message")
	rc.MsgC <- msg

}
