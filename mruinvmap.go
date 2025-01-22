package main

import "github.com/geoffreyhinton/btc_self_training/btcwire"

// MruInventoryMap provides a map that is limited to a maximum number of items
// with eviction for the oldest entry when the limit is exceeded.
type MruInventoryMap struct {
	invMap map[btcwire.InvVect]int64 // Use int64 for time for less mem.
	limit  uint
}
