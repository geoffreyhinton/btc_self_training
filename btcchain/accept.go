package btcchain

// maybeAcceptBlock potentially accepts a block into the memory block chain.
// It performs several validation checks which depend on its position within
// the block chain before adding it.  The block is expected to have already gone
// through ProcessBlock before calling this function with it.
