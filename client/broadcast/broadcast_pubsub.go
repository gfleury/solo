package broadcast

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gfleury/solo/client/broadcast/metapacket"
	"github.com/gfleury/solo/client/broadcast/prp"
	"github.com/gfleury/solo/client/crypto"
	"github.com/gfleury/solo/common/models"
	"github.com/ipfs/go-log"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Broadcaster interface {
	Lookup(dstIP string) (*models.NetworkNode, bool, bool)
	Start(ctx context.Context, host host.Host, myIP string) error
	SendPacket(ctx context.Context, packet *metapacket.MetaPacket) error
	AnnounceMyself(ctx context.Context) error
	PRPRequest(ctx context.Context, unknownDstIP string) error
}

type DefaultBroadcaster struct {
	sync.Mutex

	pubSub *pubsub.PubSub
	topic  *pubsub.Topic

	selfID peer.ID

	maxsize int
	otpKey  *crypto.OTPKey

	sealer crypto.Sealer

	logger log.StandardLogger

	PRPTable *prp.PRPTableType
}

func NewBroadcaster(
	logger log.StandardLogger,
	otpKey *crypto.OTPKey,
	maxsize int,
) Broadcaster {
	return &DefaultBroadcaster{
		otpKey:   otpKey,
		maxsize:  maxsize,
		sealer:   &crypto.DefaultSealer{},
		logger:   logger,
		PRPTable: prp.NewPRPTable(),
	}
}
func (m *DefaultBroadcaster) Lookup(dstIP string) (*models.NetworkNode, bool, bool) {
	return m.PRPTable.Lookup(dstIP)
}

func (m *DefaultBroadcaster) topicKey(salts ...string) string {
	totp := m.otpKey.TOTP(sha256.New)
	if len(salts) > 0 {
		return crypto.MD5(totp + strings.Join(salts, ":"))
	}
	return crypto.MD5(totp)
}

func (m *DefaultBroadcaster) joinBroadcast(ctxCancel context.CancelFunc) (context.CancelFunc, error) {
	var err error
	var ctx context.Context
	m.Lock()
	defer m.Unlock()

	if ctxCancel != nil {
		ctxCancel()
	}

	ctx, ctxCancel = context.WithCancel(context.Background())

	// join the broadcast room
	subscription, err := m.joinAndSubscribe()
	if err != nil {
		return ctxCancel, err
	}

	// start reading messages from the subscription in a loop
	go m.readLoop(ctx, subscription)

	return ctxCancel, nil
}

func (m *DefaultBroadcaster) Start(ctx context.Context, host host.Host, myIP string) error {
	var err error

	// Insert myself on the PRPTable
	myselfMachine := models.NewLocalNode(host, myIP)
	m.PRPTable.InsertMyselfEntry(&myselfMachine)
	m.selfID = host.ID()

	c := make(chan interface{})
	go func(c context.Context, cc chan interface{}) {
		k := ""
		for {
			select {
			default:
				currentKey := m.topicKey()
				if currentKey != k {
					k = currentKey
					cc <- nil
				}
				time.Sleep(1 * time.Second)
			case <-ctx.Done():
				close(cc)
				return
			}
		}
	}(ctx, c)

	m.logger.Debug("Creating PubGossipSub")
	// create a new PubSub service using the GossipSub router
	m.pubSub, err = pubsub.NewGossipSub(ctx, host, pubsub.WithMaxMessageSize(m.maxsize))
	if err != nil {
		return err
	}
	m.logger.Debug("Created PubGossipSub")

	var ctxCancel context.CancelFunc

	for range c {
		m.logger.Debugf("Joining new broadcast room: %s", m.topicKey())
		ctxCancel, err = m.joinBroadcast(ctxCancel)
		if err != nil {
			m.logger.Errorf("Broadcast main loop error: %s", err)
			break
		}
	}

	// Close eventual open contexts
	if ctxCancel != nil {
		ctxCancel()
	}

	return nil
}

// Publish a MetaPacket to the PubSub
func (m *DefaultBroadcaster) SendPacket(ctx context.Context, packet *metapacket.MetaPacket) error {
	bytesPacket, err := json.Marshal(packet)
	if err != nil {
		return err
	}
	return m.publishMessage(ctx, bytesPacket)
}

// Publish a raw message to the PubSub
func (m *DefaultBroadcaster) publishMessage(ctx context.Context, message []byte) error {
	m.Lock()
	defer m.Unlock()
	if m.topic != nil {
		sealedPacket, err := m.sealer.Seal(message, m.sealKey())
		if err != nil {
			return err
		}

		return m.topic.Publish(ctx, sealedPacket)
	}
	return fmt.Errorf("there is no topic ready still")
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (m *DefaultBroadcaster) readLoop(ctx context.Context, subscription *pubsub.Subscription) {
	for {
		select {
		case <-ctx.Done():
			m.Lock()
			defer m.Unlock()
			m.logger.Debug("Leaving readLoop since context is gone")
			subscription.Cancel()
			m.topic.Close()
			return
		default:
			msg, err := subscription.Next(ctx)

			if err != nil {
				return
			}

			// only forward messages delivered by others
			if msg.ReceivedFrom == m.selfID {
				continue
			}

			unsealedPacket, err := m.sealer.Unseal(msg.Data, m.sealKey())
			if err != nil {
				m.logger.Warnf("Fail to unseal receiving message %w from", err.Error())
			}

			cm := &metapacket.MetaPacket{}
			err = json.Unmarshal(unsealedPacket, cm)
			if err != nil {
				m.logger.Errorf("Unable to unmarshal received MetaPacket: %s", err)
				continue
			}

			cm.SenderID = msg.ReceivedFrom.String()

			if payload := cm.GetPayload(); payload != nil {
				replyPayload, err := payload.Process(m.logger, m.PRPTable)
				if err != nil {
					m.logger.Errorf("Unable to process received MetaPacket Payload: %s", err)
					continue
				}
				if replyPayload != nil {
					m.SendPacket(ctx, metapacket.NewFromPayload(replyPayload))
				}
			}
		}
	}
}

// connect tries to subscribe to the PubSub topic for the room name, returning
// a Room on success.
func (m *DefaultBroadcaster) joinAndSubscribe() (*pubsub.Subscription, error) {
	var err error

	m.logger.Debugf("Joining Topic: %s", m.topicKey())
	// join the pubsub topic
	m.topic, err = m.pubSub.Join(m.topicKey())
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	return m.topic.Subscribe()

}

func (m *DefaultBroadcaster) sealKey() []byte {
	return m.otpKey.TOTPSHA256(sha256.New)
}

func (m *DefaultBroadcaster) PRPRequest(ctx context.Context, unknownDstIP string) error {
	return m.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket(unknownDstIP)))
}

func (m *DefaultBroadcaster) AnnounceMyself(ctx context.Context) error {
	return m.SendPacket(ctx, metapacket.NewFromPayload(m.PRPTable.PRPReplyMyself(true)))
}
