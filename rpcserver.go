package main

import (
	"net"
	"sync"
)

// rpcServer holds the items the rpc server may need to access (config,
// shutdown, main server, etc.)
type rpcServer struct {
	started   int32
	shutdown  int32
	server    *server
	wg        sync.WaitGroup
	rpcport   string
	username  string
	password  string
	listeners []net.Listener
}
