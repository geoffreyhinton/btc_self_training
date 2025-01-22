package main

import (
	"container/list"
	"github.com/geoffreyhinton/btc_self_training/btcwire"
	"sync"
)

// txMemPool is used as a source of transactions that need to be mined into
// blocks and relayed to other peers.  It is safe for concurrent access from
// multiple peers.
type txMemPool struct {
	server        *server
	pool          map[btcwire.ShaHash]*btcwire.MsgTx
	orphans       map[btcwire.ShaHash]*btcwire.MsgTx
	orphansByPrev map[btcwire.ShaHash]*list.List
	outpoints     map[btcwire.OutPoint]*btcwire.MsgTx
	lock          sync.RWMutex
}
