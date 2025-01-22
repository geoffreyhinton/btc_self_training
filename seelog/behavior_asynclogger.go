package seelog

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
)

// MaxQueueSize is the critical number of messages in the queue that result in an immediate flush.
const (
	MaxQueueSize = 10000
)

type msgQueueItem struct {
	level   LogLevel
	context logContextInterface
	message fmt.Stringer
}

// asyncLogger represents common data for all asynchronous loggers
type asyncLogger struct {
	commonLogger
	msgQueue         *list.List
	queueHasElements *sync.Cond
}

// newAsyncLogger creates a new asynchronous logger
func newAsyncLogger(config *logConfig) *asyncLogger {
	asnLogger := new(asyncLogger)

	asnLogger.msgQueue = list.New()
	asnLogger.queueHasElements = sync.NewCond(new(sync.Mutex))

	asnLogger.commonLogger = *newCommonLogger(config, asnLogger)

	return asnLogger
}

func (asnLogger *asyncLogger) innerLog(
	level LogLevel,
	context logContextInterface,
	message fmt.Stringer) {

	asnLogger.addMsgToQueue(level, context, message)
}

func (asnLogger *asyncLogger) Close() {
	asnLogger.m.Lock()
	defer asnLogger.m.Unlock()

	if !asnLogger.closed {
		asnLogger.flushQueue(true)
		asnLogger.config.RootDispatcher.Flush()
		asnLogger.config.RootDispatcher.Close()
		asnLogger.queueHasElements.Broadcast()
	}
}

func (asnLogger *asyncLogger) Flush() {
	asnLogger.m.Lock()
	defer asnLogger.m.Unlock()

	if !asnLogger.closed {
		asnLogger.flushQueue(true)
		asnLogger.config.RootDispatcher.Flush()
	}
}

func (asnLogger *asyncLogger) flushQueue(lockNeeded bool) {

	if lockNeeded {
		asnLogger.queueHasElements.L.Lock()
		defer asnLogger.queueHasElements.L.Unlock()
	}

	for asnLogger.msgQueue.Len() > 0 {
		asnLogger.processQueueElement()
	}
}

func (asnLogger *asyncLogger) processQueueElement() {
	if asnLogger.msgQueue.Len() > 0 {
		backElement := asnLogger.msgQueue.Front()
		msg, _ := backElement.Value.(msgQueueItem)
		asnLogger.processLogMsg(msg.level, msg.message, msg.context)
		asnLogger.msgQueue.Remove(backElement)
	}
}

func (asnLogger *asyncLogger) addMsgToQueue(
	level LogLevel,
	context logContextInterface,
	message fmt.Stringer) {

	if !asnLogger.closed {
		asnLogger.queueHasElements.L.Lock()
		defer asnLogger.queueHasElements.L.Unlock()

		if asnLogger.msgQueue.Len() >= MaxQueueSize {
			fmt.Printf("Seelog queue overflow: more than %v messages in the queue. Flushing.\n", MaxQueueSize)
			asnLogger.flushQueue(false)
		}

		queueItem := msgQueueItem{level, context, message}

		asnLogger.msgQueue.PushBack(queueItem)
		asnLogger.queueHasElements.Broadcast()
	} else {
		err := errors.New(fmt.Sprintf("Queue closed! Cannot process element: %d %#v", level, message))
		reportInternalError(err)
	}
}
