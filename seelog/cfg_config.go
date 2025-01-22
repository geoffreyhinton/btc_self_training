package seelog

import (
	"bytes"
	"io"
	"os"
)

// LoggerFromConfigAsFile creates logger with config from file. File should contain valid seelog xml.
func LoggerFromConfigAsFile(fileName string) (LoggerInterface, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	conf, err := configFromReader(file)
	if err != nil {
		return nil, err
	}

	return createLoggerFromConfig(conf)
}

// LoggerFromConfigAsBytes creates a logger with config from bytes stream. Bytes should contain valid seelog xml.
func LoggerFromConfigAsBytes(data []byte) (LoggerInterface, error) {
	conf, err := configFromReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	return createLoggerFromConfig(conf)
}

// LoggerFromConfigAsString creates a logger with config from a string. String should contain valid seelog xml.
func LoggerFromConfigAsString(data string) (LoggerInterface, error) {
	return LoggerFromConfigAsBytes([]byte(data))
}

// LoggerFromWriterWithMinLevel creates a simple logger for usage with non-Seelog systems.
// Creates logger that writes to output with minimal level = minLevel.
func LoggerFromWriterWithMinLevel(output io.Writer, minLevel LogLevel) (LoggerInterface, error) {
	constraints, err := newMinMaxConstraints(minLevel, CriticalLvl)
	if err != nil {
		return nil, err
	}

	dispatcher, err := newSplitDispatcher(defaultformatter, []interface{}{output})
	if err != nil {
		return nil, err
	}

	conf, err := newConfig(constraints, make([]*logLevelException, 0), dispatcher, syncloggerTypeFromString, nil)
	if err != nil {
		return nil, err
	}

	return createLoggerFromConfig(conf)
}
