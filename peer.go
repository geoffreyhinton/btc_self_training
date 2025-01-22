package main

import (
	"container/list"
	"github.com/geoffreyhinton/btc_self_training/btcwire"
	"net"
	"sync"
	"time"
)

// peer provides a bitcoin peer for handling bitcoin communications.
type peer struct {
	server             *server
	protocolVersion    uint32
	btcnet             btcwire.BitcoinNet
	services           btcwire.ServiceFlag
	started            int32
	conn               net.Conn
	addr               string
	na                 *btcwire.NetAddress
	timeConnected      time.Time
	inbound            bool
	connected          int32
	disconnect         int32 // only to be used atomically
	persistent         bool
	versionKnown       bool
	knownAddresses     map[string]bool
	knownInventory     *MruInventoryMap
	knownInvMutex      sync.Mutex
	requestedBlocks    map[btcwire.ShaHash]bool // owned by blockmanager.
	lastBlock          int32
	retrycount         int64
	prevGetBlocksBegin *btcwire.ShaHash
	prevGetBlocksStop  *btcwire.ShaHash
	prevGetBlockMutex  sync.Mutex
	requestQueue       *list.List
	invSendQueue       *list.List
	continueHash       *btcwire.ShaHash
	outputQueue        chan btcwire.Message
	outputInvChan      chan *btcwire.InvVect
	blockProcessed     chan bool
	quit               chan bool
}
