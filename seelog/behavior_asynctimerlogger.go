package seelog

import (
	"errors"
	"time"
)

// asyncTimerLogger represents asynchronous logger which processes the log queue each
// 'duration' nanoseconds
type asyncTimerLogger struct {
	asyncLogger
	interval time.Duration
}

// newAsyncLoopLogger creates a new asynchronous loop logger
func newAsyncTimerLogger(config *logConfig, interval time.Duration) (*asyncTimerLogger, error) {

	if interval <= 0 {
		return nil, errors.New("Async logger interval should be > 0")
	}

	asnTimerLogger := new(asyncTimerLogger)

	asnTimerLogger.asyncLogger = *newAsyncLogger(config)
	asnTimerLogger.interval = interval

	go asnTimerLogger.processQueue()

	return asnTimerLogger, nil
}

func (asnTimerLogger *asyncTimerLogger) processItem() (closed bool) {
	asnTimerLogger.queueHasElements.L.Lock()
	defer asnTimerLogger.queueHasElements.L.Unlock()

	for asnTimerLogger.msgQueue.Len() == 0 && !asnTimerLogger.closed {
		asnTimerLogger.queueHasElements.Wait()
	}

	if asnTimerLogger.closed {
		return true
	}

	asnTimerLogger.processQueueElement()
	return false
}

func (asnTimerLogger *asyncTimerLogger) processQueue() {
	for !asnTimerLogger.closed {
		closed := asnTimerLogger.processItem()

		if closed {
			break
		}

		<-time.After(asnTimerLogger.interval)
	}
}
