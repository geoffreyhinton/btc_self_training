package seelog

import (
	"errors"
)

type loggerTypeFromString uint8

const (
	syncloggerTypeFromString = iota
	asyncLooploggerTypeFromString
	asyncTimerloggerTypeFromString
	adaptiveLoggerTypeFromString
	defaultloggerTypeFromString = asyncLooploggerTypeFromString
)

const (
	syncloggerTypeFromStringStr       = "sync"
	asyncloggerTypeFromStringStr      = "asyncloop"
	asyncTimerloggerTypeFromStringStr = "asynctimer"
	adaptiveLoggerTypeFromStringStr   = "adaptive"
)

// asyncTimerLoggerData represents specific data for async timer logger
type asyncTimerLoggerData struct {
	AsyncInterval uint32
}

// adaptiveLoggerData represents specific data for adaptive timer logger
type adaptiveLoggerData struct {
	MinInterval      uint32
	MaxInterval      uint32
	CriticalMsgCount uint32
}

var loggerTypeToStringRepresentations = map[loggerTypeFromString]string{
	syncloggerTypeFromString:       syncloggerTypeFromStringStr,
	asyncLooploggerTypeFromString:  asyncloggerTypeFromStringStr,
	asyncTimerloggerTypeFromString: asyncTimerloggerTypeFromStringStr,
	adaptiveLoggerTypeFromString:   adaptiveLoggerTypeFromStringStr,
}

// getLoggerTypeFromString parses a string and returns a corresponding logger type, if successful.
func getLoggerTypeFromString(logTypeString string) (level loggerTypeFromString, found bool) {
	for logType, logTypeStr := range loggerTypeToStringRepresentations {
		if logTypeStr == logTypeString {
			return logType, true
		}
	}

	return 0, false
}

// logConfig stores logging configuration. Contains messages dispatcher, allowed log level rules
// (general constraints and exceptions), and messages formats (used by nodes of dispatcher tree)
type logConfig struct {
	Constraints    logLevelConstraints  // General log level rules (>min and <max, or set of allowed levels)
	Exceptions     []*logLevelException // Exceptions to general rules for specific files or funcs
	RootDispatcher dispatcherInterface  // Root of output tree
	LogType        loggerTypeFromString
	LoggerData     interface{}
}

func newConfig(
	constraints logLevelConstraints,
	exceptions []*logLevelException,
	rootDispatcher dispatcherInterface,
	logType loggerTypeFromString,
	logData interface{}) (*logConfig, error) {
	if constraints == nil {
		return nil, errors.New("Constraints can not be nil")
	}
	if rootDispatcher == nil {
		return nil, errors.New("RootDispatcher can not be nil")
	}

	config := new(logConfig)
	config.Constraints = constraints
	config.Exceptions = exceptions
	config.RootDispatcher = rootDispatcher
	config.LogType = logType
	config.LoggerData = logData

	return config, nil
}

// IsAllowed returns true if logging with specified log level is allowed in current context.
// If any of exception patterns match current context, then exception constraints are applied. Otherwise,
// the general constraints are used.
func (config *logConfig) IsAllowed(level LogLevel, context logContextInterface) bool {
	allowed := config.Constraints.IsAllowed(level) // General rule

	// Exceptions:
	if context.IsValid() {
		for _, exception := range config.Exceptions {
			if exception.MatchesContext(context) {
				return exception.IsAllowed(level)
			}
		}
	}

	return allowed
}
