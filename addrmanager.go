package main

import (
	"container/list"
	"github.com/geoffreyhinton/btc_self_training/btcwire"
	"sync"
	"time"
)

const (
	// maxAddresses identifies the maximum number of addresses that the
	// address manager will track.
	maxAddresses = 2500

	// needAddressThreshold is the number of addresses under which the
	// address manager will claim to need more addresses.
	needAddressThreshold = 1000

	newAddressBufferSize = 50

	// dumpAddressInterval is the interval used to dump the address
	// cache to disk for future use.
	dumpAddressInterval = time.Minute * 2

	// triedBucketSize is the maximum number of addresses in each
	// tried address bucket.
	triedBucketSize = 64

	// triedBucketCount is the number of buckets we split tried
	// addresses over.
	triedBucketCount = 64

	// newBucketSize is the maximum number of addresses in each new address
	// bucket.
	newBucketSize = 64

	// newBucketCount is the number of buckets taht we spread new addresses
	// over.
	newBucketCount = 256

	// triedBucketsPerGroup is the number of trieed buckets over which an
	// address group will be spread.
	triedBucketsPerGroup = 4

	// newBucketsPerGroup is the number of new buckets over which an
	// source address group will be spread.
	newBucketsPerGroup = 32

	// newBucketsPerAddress is the number of buckets a frequently seen new
	// address may end up in.
	newBucketsPerAddress = 4

	// numMissingDays is the number of days before which we assume an
	// address has vanished if we have not seen it announced  in that long.
	numMissingDays = 30

	// numRetries is the number of tried without a single success before
	// we assume an address is bad.
	numRetries = 3

	// maxFailures is the maximum number of failures we will accept without
	// a success before considering an address bad.
	maxFailures = 10

	// minBadDays is the number of days since the last success before we
	// will consider evicting an address.
	minBadDays = 7

	// getAddrMax is the most addresses that we will send in response
	// to a getAddr (in practise the most addresses we will return from a
	// call to AddressCache()).
	getAddrMax = 2500

	// getAddrPercent is the percentage of total addresses known that we
	// will share with a call to AddressCache.
	getAddrPercent = 23

	// serialisationVersion is the current version of the on-disk format.
	serialisationVersion = 1
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

type serialisedKnownAddress struct {
	Addr        string
	Src         string
	Attempts    int
	TimeStamp   int64
	LastAttempt int64
	LastSuccess int64
	// no refcount or tried, that is available from context.
}

type serialisedAddrManager struct {
	Version      int
	Key          [32]byte
	Addresses    []*serialisedKnownAddress
	NewBuckets   [newBucketCount][]string // string is NetAddressKey
	TriedBuckets [triedBucketCount][]string
}
