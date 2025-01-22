package seelog

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	adaptiveLoggerMaxInterval         = time.Minute
	adaptiveLoggerMaxCriticalMsgCount = uint32(1000)
)

// asyncAdaptiveLogger represents asynchronous adaptive logger which acts like
// an async timer logger, but its interval depends on the current message count
// in the queue.
//
// Interval = I, minInterval = m, maxInterval = M, criticalMsgCount = C, msgCount = c:
// I = m + (C - Min(c, C)) / C * (M - m)
type asyncAdaptiveLogger struct {
	asyncLogger
	minInterval      time.Duration
	criticalMsgCount uint32
	maxInterval      time.Duration
}

// newAsyncLoopLogger creates a new asynchronous adaptive logger
func newAsyncAdaptiveLogger(
	config *logConfig,
	minInterval time.Duration,
	maxInterval time.Duration,
	criticalMsgCount uint32) (*asyncAdaptiveLogger, error) {

	if minInterval <= 0 {
		return nil, errors.New("Async adaptive logger min interval should be > 0")
	}

	if maxInterval > adaptiveLoggerMaxInterval {
		return nil, errors.New(fmt.Sprintf("Async adaptive logger max interval should be <= %s",
			adaptiveLoggerMaxInterval))
	}

	if criticalMsgCount <= 0 {
		return nil, errors.New("Async adaptive logger critical msg count should be > 0")
	}

	if criticalMsgCount > adaptiveLoggerMaxCriticalMsgCount {
		return nil, errors.New(fmt.Sprintf("Async adaptive logger critical msg count should be <= %s",
			adaptiveLoggerMaxInterval))
	}

	asnAdaptiveLogger := new(asyncAdaptiveLogger)

	asnAdaptiveLogger.asyncLogger = *newAsyncLogger(config)
	asnAdaptiveLogger.minInterval = minInterval
	asnAdaptiveLogger.maxInterval = maxInterval
	asnAdaptiveLogger.criticalMsgCount = criticalMsgCount

	go asnAdaptiveLogger.processQueue()

	return asnAdaptiveLogger, nil
}

func (asnAdaptiveLogger *asyncAdaptiveLogger) processItem() (closed bool, itemCount int) {
	asnAdaptiveLogger.queueHasElements.L.Lock()
	defer asnAdaptiveLogger.queueHasElements.L.Unlock()

	for asnAdaptiveLogger.msgQueue.Len() == 0 && !asnAdaptiveLogger.closed {
		asnAdaptiveLogger.queueHasElements.Wait()
	}

	if asnAdaptiveLogger.closed {
		return true, asnAdaptiveLogger.msgQueue.Len()
	}

	asnAdaptiveLogger.processQueueElement()
	return false, asnAdaptiveLogger.msgQueue.Len() - 1
}

// I = m + (C - Min(c, C)) / C * (M - m) =>
// I = m + cDiff * mDiff,
//
//	cDiff = (C - Min(c, C)) / C)
//	mDiff = (M - m)
func (asnAdaptiveLogger *asyncAdaptiveLogger) calcAdaptiveInterval(msgCount int) time.Duration {
	critCountF := float64(asnAdaptiveLogger.criticalMsgCount)
	cDiff := (critCountF - math.Min(float64(msgCount), critCountF)) / critCountF
	mDiff := float64(asnAdaptiveLogger.maxInterval - asnAdaptiveLogger.minInterval)

	return asnAdaptiveLogger.minInterval + time.Duration(cDiff*mDiff)
}

func (asnAdaptiveLogger *asyncAdaptiveLogger) processQueue() {
	for !asnAdaptiveLogger.closed {
		closed, itemCount := asnAdaptiveLogger.processItem()

		if closed {
			break
		}

		interval := asnAdaptiveLogger.calcAdaptiveInterval(itemCount)

		<-time.After(interval)
	}
}
