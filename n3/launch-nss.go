// launch-nss.go

package n3

import (
	"log"
	"os"
	"os/exec"
	"time"

	nats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
)

//
// handles creating & shutting down an instance of the
// nats streaming server
//

type NSS struct {
	cmd *exec.Cmd
}

//
// returns a nats streaming server instance (nss)
// which can be started and stopped
// note: this actually uses exec.Cmd to run
// a separate process for better performacne and loading
// across the avaiable processors on the host machine.
//
func NewNSS() *NSS {

	config := DefaultStanConfig("n3")
	cid := "--cid=" + config.ClusterId
	store := "--store=FILE"
	dir := "--dir=./data"
	addr := "--addr=localhost"
	port := "--port=" + config.NatsPort
	args := []string{cid, store, dir, addr, port}

	nssCmd := exec.Command("./nats-streaming-server", args...)
	nssCmd.Stdout = os.Stdout
	nssCmd.Stderr = os.Stderr
	nssCmd.Dir = "./nss"

	return &NSS{cmd: nssCmd}

}

func (nss *NSS) Start() error {
	err := nss.cmd.Start()
	if err != nil {
		return errors.Wrap(err, "unable to launch nss")
	}
	// give the nss time to come up
	time.Sleep(time.Second * 5)

	// start safety subscribers to ensure
	// topics exist
	err = startSafetySubscribers()
	if err != nil {
		return errors.Wrap(err, "could not create safety subs: ")
	}

	return nil
}

func (nss *NSS) Stop() error {

	return nss.cmd.Process.Kill()

}

//
// create durable safety subscribers for
// critical topics to ensure no message loss due to
// handshake timings
//
func startSafetySubscribers() error {

	log.Println("creating safety subscriber for nss:feed")
	sc, err := NSSConnection("feedSafety")
	if err != nil {
		return err
	}

	// main message sending routine
	go func() {
		_, err := sc.Subscribe("feed", func(m *stan.Msg) {

			// no-op

		}, stan.DurableName("feedSafety"))
		if err != nil {
			log.Println("error creating safety subscription: ", err)
			return
		}
	}()

	time.Sleep(time.Second * 3)
	log.Println("safety subscriber for nss:feed created")
	return nil

}

//
// acquire connection to the nss instance
//
func NSSConnection(clientID string) (stan.Conn, error) {

	config := DefaultStanConfig(clientID)

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
	natsOpts.Url = config.Url

	natsCon, err := natsOpts.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "nats connect failed")
	}

	var stanConn stan.Conn
	log.Printf("%#v\n", config)
	if stanConn, err = stan.Connect(
		config.ClusterId,
		config.ClientId,
		stan.NatsConn(natsCon),
	); err != nil {
		return nil, errors.Wrap(err, "stan connect failed")
	}

	return stanConn, nil
}
