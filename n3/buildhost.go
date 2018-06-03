// buildhost.go

package n3

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	ds "github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

var hostPrivKey crypto.PrivKey

// makeRoutedHost creates a LibP2P host listening on the
// given multiaddress. It will bootstrap using the
// provided PeerInfos to find nodes via ipfs discovery if --inet is true
func makeRoutedHost(listenPort int, bootstrapPeers []pstore.PeerInfo, inet bool) (host.Host, error) {

	if hostPrivKey == nil {
		var loadErr error
		privKey, loadErr := loadHostKey()
		if loadErr != nil {
			return nil, errors.Wrap(loadErr, "unable to load private key")
		}
		hostPrivKey = privKey
	}

	// Get the peer id
	pid, err := peer.IDFromPrivateKey(hostPrivKey)
	if err != nil {
		return nil, err
	}

	maddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort))
	if err != nil {
		return nil, err
	}

	// We've created the identity, now we need to store it to use the bhost constructors
	ps := pstore.NewPeerstore()
	ps.AddPrivKey(pid, hostPrivKey)
	ps.AddPubKey(pid, hostPrivKey.GetPublic())

	// Put all this together
	ctx := context.Background()
	netw, err := swarm.NewNetwork(ctx, []ma.Multiaddr{maddr}, pid, ps, nil)
	if err != nil {
		return nil, err
	}

	hostOpts := &bhost.HostOpts{
		NATManager: bhost.NewNATManager(netw),
	}

	basicHost, err := bhost.NewHost(ctx, netw, hostOpts)
	if err != nil {
		return nil, err
	}

	// // Construct a datastore (needed by the DHT). This is just a simple, in-memory thread-safe datastore.
	dstore := dsync.MutexWrap(ds.NewMapDatastore())

	// // Make the DHT
	dht := dht.NewDHT(ctx, basicHost, dstore)

	// // Make the routed host
	routedHost := rhost.Wrap(basicHost, dht)

	// connect to the ipfs nodes and get external interface
	if inet {
		err = bootstrapConnect(ctx, routedHost, bootstrapPeers)
		if err != nil {
			return nil, err
		}

		// Bootstrap the host
		err = dht.Bootstrap(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Build host multiaddress
	hostAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", routedHost.ID().Pretty()))
	if err != nil {
		log.Println("multiaddr error: ", err)
	}

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addrs := routedHost.Addrs()
	log.Println("I can be reached at:")
	for _, addr := range addrs {
		log.Println(addr.Encapsulate(hostAddr))
	}

	log.Println("Now run \"./n3 --l [port] --d [one of my addresses]\" on a different terminal")

	// return routedHost, nil
	return basicHost, nil
}

func loadHostKey() (crypto.PrivKey, error) {
	privKeyBytes, err := ioutil.ReadFile("prv.key")
	if err != nil {
		return nil, err
	}
	privK, err := crypto.UnmarshalPrivateKey(privKeyBytes)
	if err != nil {
		return nil, err
	}
	return privK, nil
}
