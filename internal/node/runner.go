package node

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amba-p2p/config"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
)

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
	go p2pNode.ConnectToPeer(ctx, cfg)

	p2pNode.WorkersCtm, err = p2pNode.JoinConsortium(ctx, networkTopic)
	if err != nil {
		log.Fatalln(fmt.Errorf("JoinConsortium: %w", err))
	}

	go PeriodicBroadcast(p2pNode, "i am message from:")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Debugln("Received signal, shutting down...")

	// shut the node down
	if err := p2pNode.Host.Close(); err != nil {
		panic(err)
	}
}

type Message struct {
	SenderID peer.ID `json:"sender_id"`
	Message  string  `json:"message"`
}
