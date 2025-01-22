package main

import (
	"container/list"
	"github.com/geoffreyhinton/btc_self_training/btcwire"
	"sync"
	"time"
)

// AddrManager provides a concurrency safe address manager for caching potential
// peers on the bitcoin network.
type AddrManager struct {
	mtx       sync.Mutex
	rand      *rand.Rand
	key       [32]byte
	addrIndex map[string]*knownAddress // address key to ka for all addrs.
	addrNew   [newBucketCount]map[string]*knownAddress
	addrTried [triedBucketCount]*list.List
	started   int32
	shutdown  int32
	wg        sync.WaitGroup
	quit      chan bool
	nTried    int
	nNew      int
}

// knownAddress tracks information about a known network address that is used
// to determine how viable an address is.
type knownAddress struct {
	na          *btcwire.NetAddress
	srcAddr     *btcwire.NetAddress
	attempts    int
	lastattempt time.Time
	lastsuccess time.Time
	tried       bool
	refs        int // reference count of new buckets
}
