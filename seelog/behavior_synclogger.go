package seelog

import (
	"fmt"
)

// syncLogger performs logging in the same goroutine where 'Trace/Debug/...'
// func was called
type syncLogger struct {
	commonLogger
}

// newSyncLogger creates a new synchronous logger
func newSyncLogger(config *logConfig) *syncLogger {
	syncLogger := new(syncLogger)

	syncLogger.commonLogger = *newCommonLogger(config, syncLogger)

	return syncLogger
}

func (cLogger *syncLogger) innerLog(
	level LogLevel,
	context logContextInterface,
	message fmt.Stringer) {

	cLogger.processLogMsg(level, message, context)
}

func (syncLogger *syncLogger) Close() {
	syncLogger.m.Lock()
	defer syncLogger.m.Unlock()

	if !syncLogger.closed {
		syncLogger.config.RootDispatcher.Close()
	}
}

func (syncLogger *syncLogger) Flush() {
	syncLogger.m.Lock()
	defer syncLogger.m.Unlock()

	if !syncLogger.closed {
		syncLogger.config.RootDispatcher.Flush()
	}
}
