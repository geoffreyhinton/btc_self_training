package main

import "time"

// config defines the configuration options for btcd.
//
// See loadConfig for details on the configuration load process.
type config struct {
	ShowVersion    bool          `short:"V" long:"version" description:"Display version information and exit"`
	ConfigFile     string        `short:"C" long:"configfile" description:"Path to configuration file"`
	DataDir        string        `short:"b" long:"datadir" description:"Directory to store data"`
	AddPeers       []string      `short:"a" long:"addpeer" description:"Add a peer to connect with at startup"`
	ConnectPeers   []string      `long:"connect" description:"Connect only to the specified peers at startup"`
	DisableListen  bool          `long:"nolisten" description:"Disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect option is used or if the --proxy option is used without the --tor option"`
	Port           string        `short:"p" long:"port" description:"Listen for connections on this port (default: 8333, testnet: 18333)"`
	MaxPeers       int           `long:"maxpeers" description:"Max number of inbound and outbound peers"`
	BanDuration    time.Duration `long:"banduration" description:"How long to ban misbehaving peers.  Valid time units are {s, m, h}.  Minimum 1 second"`
	RPCUser        string        `short:"u" long:"rpcuser" description:"Username for RPC connections"`
	RPCPass        string        `short:"P" long:"rpcpass" description:"Password for RPC connections"`
	RPCPort        string        `short:"r" long:"rpcport" description:"Listen for JSON/RPC messages on this port"`
	DisableRPC     bool          `long:"norpc" description:"Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass is specified"`
	DisableDNSSeed bool          `long:"nodnsseed" description:"Disable DNS seeding for peers"`
	Proxy          string        `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	ProxyUser      string        `long:"proxyuser" description:"Username for proxy server"`
	ProxyPass      string        `long:"proxypass" description:"Password for proxy server"`
	UseTor         bool          `long:"tor" description:"Specifies the proxy server used is a Tor node"`
	TestNet3       bool          `long:"testnet" description:"Use the test network"`
	RegressionTest bool          `long:"regtest" description:"Use the regression test network"`
	DbType         string        `long:"dbtype" description:"Database backend to use for the Block Chain"`
	Profile        string        `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	DebugLevel     string        `short:"d" long:"debuglevel" description:"Logging level {trace, debug, info, warn, error, critical}"`
}
