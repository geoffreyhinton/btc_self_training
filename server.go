package main

import (
	"github.com/geoffreyhinton/btc_self_training/btcdb"
	"github.com/geoffreyhinton/btc_self_training/btcwire"
	"net"
	"sync"
	"time"
)

const supportedServices = btcwire.SFNodeNetwork
const connectionRetryInterval = time.Second * 10

// defaultMaxOutbound is the default number of max outbound peers.
const defaultMaxOutbound = 8

// broadcastMsg provides the ability to house a bitcoin message to be broadcast
// to all connected peers except specified excluded peers.
type broadcastMsg struct {
	message      btcwire.Message
	excludePeers []*peer
}

// server provides a bitcoin server for handling communications to and from
// bitcoin peers.
type server struct {
	nonce         uint64
	listeners     []net.Listener
	btcnet        btcwire.BitcoinNet
	started       int32 // atomic
	shutdown      int32 // atomic
	shutdownSched int32 // atomic
	addrManager   *AddrManager
	rpcServer     *rpcServer
	blockManager  *blockManager
	txMemPool     *txMemPool
	newPeers      chan *peer
	donePeers     chan *peer
	banPeers      chan *peer
	wakeup        chan bool
	relayInv      chan *btcwire.InvVect
	broadcast     chan broadcastMsg
	wg            sync.WaitGroup
	quit          chan bool
	db            btcdb.Db
}
