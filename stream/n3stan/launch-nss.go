// launch-nss.go

package n3stan

import (
	"os"
	"os/exec"

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
	return nil
}

func (nss *NSS) Stop() error {

	return nss.cmd.Process.Kill()

}
