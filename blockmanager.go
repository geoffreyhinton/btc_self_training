package main

import (
	"github.com/geoffreyhinton/btc_self_training/btcchain"
	"github.com/geoffreyhinton/btc_self_training/btcwire"
	"sync"
	"time"
)

// blockManager provides a concurrency safe block manager for handling all
// incoming blocks.
type blockManager struct {
	server            *server
	started           int32
	shutdown          int32
	blockChain        *btcchain.BlockChain
	blockPeer         map[btcwire.ShaHash]*peer
	requestedBlocks   map[btcwire.ShaHash]bool
	receivedLogBlocks int64
	receivedLogTx     int64
	lastBlockLogTime  time.Time
	processingReqs    bool
	syncPeer          *peer
	msgChan           chan interface{}
	wg                sync.WaitGroup
	quit              chan bool
}
