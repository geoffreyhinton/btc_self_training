package seelog

// asyncLoopLogger represents asynchronous logger which processes the log queue in
// a 'for' loop
type asyncLoopLogger struct {
	asyncLogger
}

// newAsyncLoopLogger creates a new asynchronous loop logger
func newAsyncLoopLogger(config *logConfig) *asyncLoopLogger {

	asnLoopLogger := new(asyncLoopLogger)

	asnLoopLogger.asyncLogger = *newAsyncLogger(config)

	go asnLoopLogger.processQueue()

	return asnLoopLogger
}

func (asnLoopLogger *asyncLoopLogger) processItem() (closed bool) {
	asnLoopLogger.queueHasElements.L.Lock()
	defer asnLoopLogger.queueHasElements.L.Unlock()

	for asnLoopLogger.msgQueue.Len() == 0 && !asnLoopLogger.closed {
		asnLoopLogger.queueHasElements.Wait()
	}

	if asnLoopLogger.closed {
		return true
	}

	asnLoopLogger.processQueueElement()
	return false
}

func (asnLoopLogger *asyncLoopLogger) processQueue() {
	for !asnLoopLogger.closed {
		closed := asnLoopLogger.processItem()

		if closed {
			break
		}
	}
}
