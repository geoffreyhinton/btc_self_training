package main

import (
	"github.com/geoffreyhinton/btc_self_training/btcwire"
	"math/big"
)

// params is used to group parameters for various networks such as the main
// network and test networks.
type params struct {
	netName      string
	btcnet       btcwire.BitcoinNet
	genesisBlock *btcwire.MsgBlock
	genesisHash  *btcwire.ShaHash
	powLimit     *big.Int
	powLimitBits uint32
	peerPort     string
	listenPort   string
	rpcPort      string
	dnsSeeds     []string
}
