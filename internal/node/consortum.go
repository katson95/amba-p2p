package node

import (
	"context"
	"encoding/json"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
)

type Consortium struct {
	ctx          context.Context
	selfId       peer.ID
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
	Ps           *pubsub.PubSub
	Messages     chan *Message
	Headers      []string
}

func (c *Consortium) Publish(msg *Message) error {

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = c.topic.Publish(c.ctx, msgBytes)
	if err != nil {
		return err
	}
	log.Printf("-Published Message(%d chars): %s\n", len(msgBytes), string(msgBytes))
	return nil
}
