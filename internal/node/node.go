package node

import (
	"context"
	"fmt"
	mrand "math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	config "github.com/amba-p2p/config"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	routingDiscovery "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

const ProtocolID = "/i360/1.0.0"

type Node struct {
	Host       host.Host
	KadDHT     *dht.IpfsDHT
	PubSub     *pubsub.PubSub
	MastersCtm *Consortium
	WorkersCtm *Consortium
}

type Consortium struct {
	ctx          context.Context
	selfId       peer.ID
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
	Ps           *pubsub.PubSub
	Messages     chan *Message
	Headers      []string
}

type Message struct {
	SenderID peer.ID `json:"sender_id"`
	Message  string  `json:"message"`
}

func New(ctx context.Context, cfg config.Config) (*Node, error) {
	//---1. Create private key---
	privateKey, _, err := crypto.GenerateKeyPairWithReader(
		crypto.RSA,
		2048,
		mrand.New(mrand.NewSource(cfg.Seed)),
	)
	if err != nil {
		return nil, err
	}

	//---2. Create listening address---
	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.Port))
	if err != nil {
		return nil, err
	}

	//---3. Setup Limiter config ---
	currentDir, _ := os.Getwd()

	limiterCfg, err := os.Open(currentDir + "/internal/node/limiterCfg.json")
	if err != nil {
		panic(err)
	}
	limiter, err := rcmgr.NewDefaultLimiterFromJSON(limiterCfg)
	if err != nil {
		panic(err)
	}
	rcm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		panic(err)
	}

	//---4. Create node options ---
	opts := []libp2p.Option{
		libp2p.ListenAddrs(addr),
		libp2p.Identity(privateKey),
		libp2p.ResourceManager(rcm),
		libp2p.DefaultTransports,
		libp2p.NATPortMap(),
		libp2p.Security(tls.ID, tls.New),
		libp2p.DefaultEnableRelay,
		libp2p.DefaultConnectionManager,
		libp2p.DefaultMuxers,
		libp2p.EnableNATService(),
	}
	p2pHost, err := libp2p.New(opts...)
	if err != nil {
		log.Fatalln(fmt.Errorf("libp2p.New: %w", err))
	}

	//---5.Set pubsub---
	ps, err := pubsub.NewGossipSub(ctx, p2pHost)
	if err != nil {
		log.Fatalln(fmt.Errorf("pubsub.NewGossipSub: %w", err))
	}

	//---6.Create kadDHT---
	kadDHT, err := dht.New(ctx, p2pHost)
	if err != nil {
		log.Fatalln(fmt.Errorf("dht.New: %w", err))
	}
	log.Println("bootstrapping the DHT")
	if err = kadDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("kadDHT.Bootstrap: %w", err)
	}
	var wg sync.WaitGroup
	for _, peerAddr := range cfg.BootstrapPeers {
		peerInfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := p2pHost.Connect(ctx, *peerInfo); err != nil {
				log.Warn(fmt.Errorf("p2pHost.Connect: %w", err))
			} else {
				log.Info("Connection established with bootstrap node:", *peerInfo)
			}
		}()
	}
	wg.Wait()

	return &Node{
		Host:   p2pHost,
		KadDHT: kadDHT,
		PubSub: ps,
	}, nil
}

func Run() {

	ctx := context.Background()

	cfg, err := config.ParseFlags()
	if err != nil {
		log.Fatalln(fmt.Errorf("cfg.ParseFlags: %w", err))
	}

	p2pNode, err := New(ctx, cfg)
	if err != nil {

	}

	log.Println("##################################################################")
	log.Println("I am:", p2pNode.Host.ID())
	for _, addr := range p2pNode.Host.Addrs() {
		fmt.Printf("%s: %s/p2p/%s", "I am @:", addr, p2pNode.Host.ID().String())
		fmt.Println()
	}
	log.Println("##################################################################")

	go AdvertiseSelf(ctx, p2pNode, cfg)
	go ConnectToPeers(ctx, p2pNode, cfg)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Debugln("Received signal, shutting down...")

	// shut the node down
	if err := p2pNode.Host.Close(); err != nil {
		panic(err)
	}
}

func AdvertiseSelf(ctx context.Context, n *Node, cfg config.Config) {

	routingDsc := routingDiscovery.NewRoutingDiscovery(n.KadDHT)

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	for {
		log.Println("Announcing self...")
		dutil.Advertise(ctx, routingDsc, cfg.Rendezvous)
		time.Sleep(20 * time.Second)
	}
}

func ConnectToPeers(ctx context.Context, n *Node, cfg config.Config) {

	routingDsc := routingDiscovery.NewRoutingDiscovery(n.KadDHT)

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	for {
		log.Println("Searching for peerAddrInfors...")
		peerInfos, err := routingDsc.FindPeers(ctx, cfg.Rendezvous)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
				"type":  "FindPeers",
			}).Fatalln("error finding peerAddrInfors!")
		}
		for peerInfo := range peerInfos {
			if len(peerInfo.Addrs) > 0 {
				if peerInfo.ID == n.Host.ID() {
					continue
				}
				//n.Host.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)

			}
			if len(peerInfo.Addrs) > 0 {
				status := n.Host.Network().Connectedness(peerInfo.ID)
				if status == network.CanConnect || status == network.NotConnected {
					if err := n.Host.Connect(ctx, peerInfo); err != nil {
						n.Host.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)
						fmt.Printf("error connecting to peer %s: %s\n", peerInfo.ID.String(), err)
					} else {
						fmt.Printf("########################################\n")
						log.Println("connected to peer: ", peerInfo.ID)
						fmt.Printf("########################################\n")
					}
				}

			}
		}
		time.Sleep(2 * time.Second)
	}
}
