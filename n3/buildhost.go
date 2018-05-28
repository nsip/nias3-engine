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

// // can be useful if you want non-persistent identities
// // especially for testing
// func makeRandomHost(port int) host.Host {
// 	// Ignoring most errors for brevity
// 	// See echo example for more details and better implementation
// 	priv, pub, _ := crypto.GenerateKeyPair(crypto.RSA, 2048)
// 	pid, _ := peer.IDFromPublicKey(pub)
// 	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
// 	ps := pstore.NewPeerstore()
// 	ps.AddPrivKey(pid, priv)
// 	ps.AddPubKey(pid, pub)
// 	n, _ := swarm.NewNetwork(context.Background(),
// 		[]ma.Multiaddr{listen}, pid, ps, nil)
// 	return bhost.New(n)
// }

// // makeBasicHost creates a LibP2P host listening on the
// // given multiaddress.
// //
// // use this rather than routedHost when on known networks or where
// // sharig ip address is ok,
// func makeBasicHost(listenPort int) (host.Host, error) {

// 	if hostPrivKey == nil {
// 		var loadErr error
// 		privKey, loadErr := loadHostKey()
// 		if loadErr != nil {
// 			r := rand.Reader
// 			var genErr error
// 			privKey, _, genErr = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
// 			if genErr != nil {
// 				return nil, errors.Wrap(genErr, "unable to generate host identity key")
// 			}
// 		}
// 		hostPrivKey = privKey
// 	}

// 	err := saveHostKey(hostPrivKey)
// 	if err != nil {
// 		log.Println("cannot persist host id, id is ephemoral")
// 	}

// 	opts := []libp2p.Option{
// 		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
// 		libp2p.Identity(hostPrivKey),
// 	}

// 	basicHost, err := libp2p.New(context.Background(), opts...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Build host multiaddress
// 	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

// 	// Now we can build a full multiaddress to reach this host
// 	// by encapsulating both addresses:
// 	addrs := basicHost.Addrs()
// 	// fullAddr := addr.Encapsulate(hostAddr)
// 	// log.Printf("I am %s\n", fullAddr)
// 	// log.Println("I can be reached at:")
// 	for _, addr := range addrs {
// 		log.Println(addr.Encapsulate(hostAddr))
// 	}

// 	log.Println("Now run ./n3 -d [one of my addresses] on a different terminal")

// 	return basicHost, nil
// }

// makeRoutedHost creates a LibP2P host listening on the
// given multiaddress. It will bootstrap using the
// provided PeerInfos to find nodes via ipfs discovery
func makeRoutedHost(listenPort int, bootstrapPeers []pstore.PeerInfo, globalFlag string) (host.Host, error) {

	if hostPrivKey == nil {
		var loadErr error
		privKey, loadErr := loadHostKey()
		if loadErr != nil {
			// r := rand.Reader
			// var genErr error
			// privKey, _, genErr = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
			// if genErr != nil {
			// 	return nil, errors.Wrap(genErr, "unable to generate host identity key")
			// }
			return nil, errors.Wrap(loadErr, "unable to load private key")
		}
		hostPrivKey = privKey
	}

	// err := saveHostKey(hostPrivKey)
	// if err != nil {
	// 	log.Println("cannot persist host id, id is ephemoral")
	// }

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

	// Construct a datastore (needed by the DHT). This is just a simple, in-memory thread-safe datastore.
	dstore := dsync.MutexWrap(ds.NewMapDatastore())

	// Make the DHT
	dht := dht.NewDHT(ctx, basicHost, dstore)

	// Make the routed host
	routedHost := rhost.Wrap(basicHost, dht)

	// connect to the chosen ipfs nodes
	err = bootstrapConnect(ctx, routedHost, bootstrapPeers)
	if err != nil {
		return nil, err
	}

	// Bootstrap the host
	err = dht.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", routedHost.ID().Pretty()))
	if err != nil {
		log.Println("multiaddr error: ", err)
	}

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	// addr := routedHost.Addrs()[0]
	addrs := routedHost.Addrs()
	log.Println("I can be reached at:")
	for _, addr := range addrs {
		log.Println(addr.Encapsulate(hostAddr))
	}

	log.Printf("Now run \"./n3 -d %s%s\" on a different terminal\n", routedHost.ID().Pretty(), globalFlag)

	return routedHost, nil
}

// func saveHostKey(privKey crypto.PrivKey) error {
// 	privKeyBytes, err := crypto.MarshalPrivateKey(privKey)
// 	if err != nil {
// 		return err
// 	}
// 	err = ioutil.WriteFile("hostprv.key", privKeyBytes, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

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
