package main

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/geoffreyhinton/btc_self_training/btcwire"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
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

// NetAddressKey returns a string key in the form of ip:port for IPv4 addresses
// or [ip]:port for IPv6 addresses.
func NetAddressKey(na *btcwire.NetAddress) string {
	port := strconv.FormatUint(uint64(na.Port), 10)
	addr := net.JoinHostPort(na.IP.String(), port)
	return addr
}

// bad returns true if the address in question has not been tried in the last
// minute and meets one of the following criteria:
// 1) It claims to be from the future
// 2) It hasn't been seen in over a month
// 3) It has failed at least three times and never succeeded
// 4) It has failed ten times in the last week
// All addresses that meet these criteria are assumed to be worthless and not
// worth keeping hold of.
func bad(ka *knownAddress) bool {
	if ka.lastattempt.After(time.Now().Add(-1 * time.Minute)) {
		return false
	}

	// From the future?
	if ka.na.Timestamp.After(time.Now().Add(10 * time.Minute)) {
		return true
	}

	// Over a month old?
	if ka.na.Timestamp.After(time.Now().Add(-1 * numMissingDays * time.Hour * 24)) {
		return true
	}

	// Never succeeded?
	if ka.lastsuccess.IsZero() && ka.attempts >= numRetries {
		return true
	}

	// Hasn't succeeded in too long?
	if !ka.lastsuccess.After(time.Now().Add(-1*minBadDays*time.Hour*24)) &&
		ka.attempts >= maxFailures {
		return true
	}

	return false
}

// expireNew makes space in the new buckets by expiring the really bad entries.
// If no bad entries are available we look at a few and remove the oldest.
func (a *AddrManager) expireNew(bucket int) {
	// First see if there are any entries that are so bad we can just throw
	// them away. otherwise we throw away the oldest entry in the cache.
	// Bitcoind here chooses four random and just throws the oldest of
	// those away, but we keep track of oldest in the initial traversal and
	// use that information instead.
	var oldest *knownAddress
	for k, v := range a.addrNew[bucket] {
		if bad(v) {
			log.Tracef("[AMGR] expiring bad address %v", k)
			delete(a.addrNew[bucket], k)
			a.nNew--
			v.refs--
			if v.refs == 0 {
				delete(a.addrIndex, k)
			}
			return
		}
		if oldest == nil {
			oldest = v
		} else if !v.na.Timestamp.After(oldest.na.Timestamp) {
			oldest = v
		}
	}

	if oldest != nil {
		key := NetAddressKey(oldest.na)
		log.Tracef("[AMGR] expiring oldest address %v", key)

		delete(a.addrNew[bucket], key)
		a.nNew--
		oldest.refs--
		if oldest.refs == 0 {
			delete(a.addrIndex, key)
		}
	}
}

// savePeers saves all the known addresses to a file so they can be read back
// in at next run.
func (a *AddrManager) savePeers() {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	// First we make a serialisable datastructure so we can encode it to
	// json.

	sam := new(serialisedAddrManager)
	sam.Version = serialisationVersion
	copy(sam.Key[:], a.key[:])

	sam.Addresses = make([]*serialisedKnownAddress, len(a.addrIndex))
	i := 0
	for k, v := range a.addrIndex {
		ska := new(serialisedKnownAddress)
		ska.Addr = k
		ska.TimeStamp = v.na.Timestamp.Unix()
		ska.Src = NetAddressKey(v.srcAddr)
		ska.Attempts = v.attempts
		ska.LastAttempt = v.lastattempt.Unix()
		ska.LastSuccess = v.lastsuccess.Unix()
		// Tried and refs are implicit in the rest of the structure
		// and will be worked out from context on unserialisation.
		sam.Addresses[i] = ska
		i++
	}
	for i := range a.addrNew {
		sam.NewBuckets[i] = make([]string, len(a.addrNew[i]))
		j := 0
		for k := range a.addrNew[i] {
			sam.NewBuckets[i][j] = k
			j++
		}
	}
	for i := range a.addrTried {
		sam.TriedBuckets[i] = make([]string, a.addrTried[i].Len())
		j := 0
		for e := a.addrTried[i].Front(); e != nil; e = e.Next() {
			ka := e.Value.(*knownAddress)
			sam.TriedBuckets[i][j] = NetAddressKey(ka.na)
			j++
		}
	}

	// May give some way to specify this later.
	filename := "peers.json"
	filePath := filepath.Join(cfg.DataDir, filename)

	w, err := os.Create(filePath)
	if err != nil {
		log.Error("Error opening file: ", filePath, err)
	}
	enc := json.NewEncoder(w)
	defer w.Close()
	enc.Encode(&sam)
}

func deserialiseNetAddress(addr string) (*btcwire.NetAddress, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, err
	}
	na := btcwire.NewNetAddressIPPort(ip, uint16(port),
		btcwire.SFNodeNetwork)
	return na, nil
}

func (a *AddrManager) deserialisePeers(filePath string) error {

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	r, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("%s error opening file: %v", filePath, err)
	}
	defer r.Close()

	var sam serialisedAddrManager
	dec := json.NewDecoder(r)
	err = dec.Decode(&sam)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", filePath, err)
	}

	if sam.Version != serialisationVersion {
		return fmt.Errorf("unknown version %v in serialised "+
			"addrmanager", sam.Version)
	}
	copy(a.key[:], sam.Key[:])

	for _, v := range sam.Addresses {
		ka := new(knownAddress)
		ka.na, err = deserialiseNetAddress(v.Addr)
		if err != nil {
			return fmt.Errorf("failed to deserialise netaddress "+
				"%s: %v", v.Addr, err)
		}
		ka.srcAddr, err = deserialiseNetAddress(v.Src)
		if err != nil {
			return fmt.Errorf("failed to deserialise netaddress "+
				"%s: %v", v.Src, err)
		}
		ka.attempts = v.Attempts
		ka.lastattempt = time.Unix(v.LastAttempt, 0)
		ka.lastsuccess = time.Unix(v.LastSuccess, 0)
		a.addrIndex[NetAddressKey(ka.na)] = ka
	}

	for i := range sam.NewBuckets {
		for _, val := range sam.NewBuckets[i] {
			ka, ok := a.addrIndex[val]
			if !ok {
				return fmt.Errorf("newbucket contains %s but "+
					"none in address list", val)
			}

			if ka.refs == 0 {
				a.nNew++
			}
			ka.refs++
			a.addrNew[i][val] = ka
		}
	}
	for i := range sam.TriedBuckets {
		for _, val := range sam.TriedBuckets[i] {
			ka, ok := a.addrIndex[val]
			if !ok {
				return fmt.Errorf("Newbucket contains %s but "+
					"none in address list", val)
			}

			ka.tried = true
			a.nTried++
			a.addrTried[i].PushBack(ka)
		}
	}

	// Sanity checking.
	for k, v := range a.addrIndex {
		if v.refs == 0 && !v.tried {
			return fmt.Errorf("address %s after serialisation "+
				"with no references", k)
		}

		if v.refs > 0 && v.tried {
			return fmt.Errorf("address %s after serialisation "+
				"which is both new and tried!", k)
		}
	}

	return nil
}

func (a *AddrManager) getNewBucket(netAddr, srcAddr *btcwire.NetAddress) int {
	// bitcoind:
	// doublesha256(key + sourcegroup + int64(doublesha256(key + group + sourcegroup))%bucket_per_source_group) % num_new_buckes

	data1 := []byte{}
	data1 = append(data1, a.key[:]...)
	data1 = append(data1, []byte(GroupKey(netAddr))...)
	data1 = append(data1, []byte(GroupKey(srcAddr))...)
	hash1 := btcwire.DoubleSha256(data1)
	hash64 := binary.LittleEndian.Uint64(hash1)
	hash64 %= newBucketsPerGroup
	hashbuf := new(bytes.Buffer)
	binary.Write(hashbuf, binary.LittleEndian, hash64)
	data2 := []byte{}
	data2 = append(data2, a.key[:]...)
	data2 = append(data2, GroupKey(srcAddr)...)
	data2 = append(data2, hashbuf.Bytes()...)

	hash2 := btcwire.DoubleSha256(data2)
	return int(binary.LittleEndian.Uint64(hash2) % newBucketCount)
}

// RFC1918: IPv4 Private networks (10.0.0.0/8, 192.168.0.0/16, 172.16.0.0/12)
var rfc1918ten = net.IPNet{IP: net.ParseIP("10.0.0.0"),
	Mask: net.CIDRMask(8, 32)}
var rfc1918oneninetwo = net.IPNet{IP: net.ParseIP("192.168.0.0"),
	Mask: net.CIDRMask(16, 32)}
var rfc1918oneseventwo = net.IPNet{IP: net.ParseIP("172.16.0.0"),
	Mask: net.CIDRMask(12, 32)}

func RFC1918(na *btcwire.NetAddress) bool {
	return rfc1918ten.Contains(na.IP) ||
		rfc1918oneninetwo.Contains(na.IP) ||
		rfc1918oneseventwo.Contains(na.IP)
}

// RFC3849 IPv6 Documentation address  (2001:0DB8::/32)
var rfc3849 = net.IPNet{IP: net.ParseIP("2001:0DB8::"),
	Mask: net.CIDRMask(32, 128)}

func RFC3849(na *btcwire.NetAddress) bool {
	return rfc3849.Contains(na.IP)
}

// RFC3927 IPv4 Autoconfig (169.254.0.0/16)
var rfc3927 = net.IPNet{IP: net.ParseIP("169.254.0.0"), Mask: net.CIDRMask(16, 32)}

func RFC3927(na *btcwire.NetAddress) bool {
	return rfc3927.Contains(na.IP)
}

// RFC3964 IPv6 6to4 (2002::/16)
var rfc3964 = net.IPNet{IP: net.ParseIP("2002::"),
	Mask: net.CIDRMask(16, 128)}

func RFC3964(na *btcwire.NetAddress) bool {
	return rfc3964.Contains(na.IP)
}

// RFC4193 IPv6 unique local (FC00::/7)
var rfc4193 = net.IPNet{IP: net.ParseIP("FC00::"),
	Mask: net.CIDRMask(7, 128)}

func RFC4193(na *btcwire.NetAddress) bool {
	return rfc4193.Contains(na.IP)
}

// RFC4380 IPv6 Teredo tunneling (2001::/32)
var rfc4380 = net.IPNet{IP: net.ParseIP("2001::"),
	Mask: net.CIDRMask(32, 128)}

func RFC4380(na *btcwire.NetAddress) bool {
	return rfc4380.Contains(na.IP)
}

// RFC4843 IPv6 ORCHID: (2001:10::/28)
var rfc4843 = net.IPNet{IP: net.ParseIP("2001:10::"),
	Mask: net.CIDRMask(28, 128)}

func RFC4843(na *btcwire.NetAddress) bool {
	return rfc4843.Contains(na.IP)
}

// RFC4862 IPv6 Autoconfig (FE80::/64)
var rfc4862 = net.IPNet{IP: net.ParseIP("FE80::"),
	Mask: net.CIDRMask(64, 128)}

func RFC4862(na *btcwire.NetAddress) bool {
	return rfc4862.Contains(na.IP)
}

// RFC6052: IPv6 well known prefix (64:FF9B::/96)
var rfc6052 = net.IPNet{IP: net.ParseIP("64:FF9B::"),
	Mask: net.CIDRMask(96, 128)}

func RFC6052(na *btcwire.NetAddress) bool {
	return rfc6052.Contains(na.IP)
}

// RFC6145: IPv6 IPv4 translated address ::FFFF:0:0:0/96
var rfc6145 = net.IPNet{IP: net.ParseIP("::FFFF:0:0:0"),
	Mask: net.CIDRMask(96, 128)}

func RFC6145(na *btcwire.NetAddress) bool {
	return rfc6145.Contains(na.IP)
}

func Tor(na *btcwire.NetAddress) bool {
	// bitcoind encodes a .onion address as a 16 byte number by decoding the
	// address prior to the .onion (i.e. the key hash) base32 into a ten
	// byte number. it then stores the first 6 bytes of the address as
	// 0xfD, 0x87, 0xD8, 0x7e, 0xeb, 0x43
	// making the format
	// { magic 6 bytes, 10 bytes base32 decode of key hash }
	// Since we use btcwire.NetAddress to represent and address we may
	// well have to emulate this.
	// XXX fillmein
	return false
}

var zero4 = net.IPNet{IP: net.ParseIP("0.0.0.0"),
	Mask: net.CIDRMask(8, 32)}

func Local(na *btcwire.NetAddress) bool {
	return na.IP.IsLoopback() || zero4.Contains(na.IP)
}

// Valid returns true if an address is not one of the invalid formats.
// For IPv4 these are either a 0 or all bits set address. For IPv6 a zero
// address or one that matches the RFC3849 documentation address format.
func Valid(na *btcwire.NetAddress) bool {
	// IsUnspecified returns if address is 0, so only all bits set, and
	// RFC3849 need to be explicitly checked. bitcoind here also checks for
	// invalid protocol addresses from earlier versions of bitcoind (before
	// 0.2.9), however, since protocol versions before 70001 are
	// disconnected by the bitcoin network now we have elided it.
	return !(na.IP.IsUnspecified() || RFC3849(na) ||
		na.IP.Equal(net.IPv4bcast))
}

// Routable returns whether a netaddress is routable on the public internet or
// not. This is true as long as the address is valid and is not in any reserved
// ranges.
func Routable(na *btcwire.NetAddress) bool {
	return Valid(na) && !(RFC1918(na) || RFC3927(na) || RFC4862(na) ||
		RFC4193(na) || Tor(na) || RFC4843(na) || Local(na))
}

// GroupKey returns a string representing the network group an address
// is part of.
// This is the /16 for IPv6, the /32 (/36 for he.net) for IPv6, the string
// "local" for a local address and the string "unroutable for an unroutable
// address.
func GroupKey(na *btcwire.NetAddress) string {
	if Local(na) {
		return "local"
	}
	if !Routable(na) {
		return "unroutable"
	}

	if ipv4 := na.IP.To4(); ipv4 != nil {
		return (&net.IPNet{IP: na.IP, Mask: net.CIDRMask(16, 32)}).String()
	}
	if RFC6145(na) || RFC6052(na) {
		// last four bytes are the ip address
		ip := net.IP(na.IP[12:16])
		return (&net.IPNet{IP: ip, Mask: net.CIDRMask(16, 32)}).String()
	}

	if RFC3964(na) {
		ip := net.IP(na.IP[2:7])
		return (&net.IPNet{IP: ip, Mask: net.CIDRMask(16, 32)}).String()

	}
	if RFC4380(na) {
		// teredo tunnels have the last 4 bytes as the v4 address XOR
		// 0xff.
		ip := net.IP(make([]byte, 4))
		for i, byte := range na.IP[12:16] {
			ip[i] = byte ^ 0xff
		}
		return (&net.IPNet{IP: ip, Mask: net.CIDRMask(16, 32)}).String()
	}
	// XXX tor?
	if Tor(na) {
		panic("oga should have implemented me")
	}

	// OK, so now we know ourselves to be a IPv6 address.
	// bitcoind uses /32 for everything but what it calls he.net, which is
	// it uses /36 for. he.net is actualy 2001:470::/32, whereas bitcoind
	// counts it as 2011:470::/32.

	bits := 32
	heNet := &net.IPNet{IP: net.ParseIP("2011:470::"),
		Mask: net.CIDRMask(32, 128)}
	if heNet.Contains(na.IP) {
		bits = 36
	}

	return (&net.IPNet{IP: na.IP, Mask: net.CIDRMask(bits, 128)}).String()
}
