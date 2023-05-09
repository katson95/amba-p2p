package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	config "github.com/amba-p2p/config"
	"github.com/amba-p2p/internal/platform/locator"
	"github.com/amba-p2p/pkg/http/rest/client"
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
	"github.com/spf13/viper"
)

const networkTopic = "/i360/1.0.0"
const consensusTopic = "/i360/master/1.0.0"

type Node struct {
	Host       host.Host
	KadDHT     *dht.IpfsDHT
	PubSub     *pubsub.PubSub
	MastersCtm *Consortium
	WorkersCtm *Consortium
	Locator    *locator.Locator
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

	//currentDir, _ := os.Getwd()
	//---3. Setup Limiter config ---
	viper.AddConfigPath("./internal/config")
	viper.SetConfigName("limiterCfg") // Register config file name (no extension)
	viper.SetConfigType("json")       // Look for specific type
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	/*port := viper.Get("System.StreamsInbound")

	fmt.Println(port)

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(f)
	if err != nil {
		log.Fatal(err)
	}
	limiterCfg, err := os.ReadFile("/home/limiterCfg.json")
	if err != nil {
		panic(err)
	}
		limiterCfg, err := os.Open(currentDir + "/internal/node/limiterCfg.json")
		if err != nil {
			panic(err)
		}*/
	limiterCfg, err := os.ReadFile("./internal/config/limiterCfg.json")
	limiter, err := rcmgr.NewDefaultLimiterFromJSON(bytes.NewReader(limiterCfg))
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

	//---Initialize REST Client---
	var geoLocator *locator.Locator

	locatorClient := client.NewClient("https://ipinfo.io", 12000)
	if err := locatorClient.ExecuteRequest(ctx, http.MethodGet, "", "/json", &geoLocator, nil); err != nil {
		return nil, err
	}

	return &Node{
		Host:    p2pHost,
		KadDHT:  kadDHT,
		PubSub:  ps,
		Locator: geoLocator,
	}, nil
}

func (n *Node) ConnectToPeer(ctx context.Context, cfg config.Config) {

	routingDsc := routingDiscovery.NewRoutingDiscovery(n.KadDHT)

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	for {
		log.Println("Searching for peers...")
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

func (n *Node) JoinConsortium(ctx context.Context, topicID string) (*Consortium, error) {
	topic, err := n.PubSub.Join(topicID)

	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	consortium := &Consortium{
		ctx:          ctx,
		subscription: sub,
		selfId:       n.Host.ID(),
		topic:        topic,
		Ps:           n.PubSub,
		Messages:     make(chan *Message, 1024),
		Headers:      make([]string, 0),
	}

	go n.SubscribeLoop()

	return consortium, nil
}

func (n *Node) SubscribeLoop() {
	for {
		inboundMsg, err := n.WorkersCtm.subscription.Next(n.WorkersCtm.ctx)

		if err != nil {
			close(n.WorkersCtm.Messages)
			return
		}

		if inboundMsg.ReceivedFrom == n.WorkersCtm.selfId {
			continue
		}

		msg := &Message{}
		err = json.Unmarshal(inboundMsg.Data, msg)
		if err != nil {
			continue
		}

		log.Infoln("Received message:", msg)

		//---1. If i am master, read network status of sender, or contact masters of sender's zone of operation---

		//---2. Master in senders zone will computer federation requirements and send joining information to sender

		n.WorkersCtm.Messages <- msg
	}
}

func (n *Node) Broadcast(msg string) {

	message := Message{
		Message:  msg,
		SenderID: n.Host.ID(),
	}

	if err := n.WorkersCtm.Publish(&message); err != nil {
		log.Println("- Error publishing to network:", err)
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

func PeriodicBroadcast(n *Node, msg string) {

	//--1. Fetch my network zone---
	zone := n.Locator.Region
	fmt.Println(zone)

	//--2. Announce my location to network consortium---

	//--3. Wait for response from a master already in network/masters consortium (Retry x times)
	//--3.1 Response: my role in the cluster (including master(s) or k8s cluster i should join & joining key)

	//--4. In case of a response, run k8s command to join existing cluster

	//--5. In case of no response, assume the role of a master
	//--5.1 Establish new cluster state, write state to the database
	//--5.2 Wait/Listen for new nodes to join, broadcasting their location (zone of operation) to the network consortium
	//--5.3 Assign role of new joiner as worker/master based on federation status i.e. the number of masters/workers

	/*	for {
		time.Sleep(time.Second * 15)
		peers := n.WorkersCtm.Ps.ListPeers(networkTopic)
		if len(peers) != 0 {
			log.Printf("- Found %d other peers in the network: %s\n", len(peers), peers)
			log.Info("Sending message")
			n.Broadcast(msg + "" + n.Host.ID().String())
		} else {
			log.Warn("- Found 0 other peers in the network: \n")
		}

	}*/

}
