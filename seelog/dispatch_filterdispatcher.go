package seelog

import (
	"fmt"
)

// A filterDispatcher writes the given message to underlying receivers only if message log level
// is in the allowed list.
type filterDispatcher struct {
	*dispatcher
	allowList map[LogLevel]bool
}

// newFilterDispatcher creates a new filterDispatcher using a list of allowed levels.
func newFilterDispatcher(formatter *formatter, receivers []interface{}, allowList ...LogLevel) (*filterDispatcher, error) {
	disp, err := createDispatcher(formatter, receivers)
	if err != nil {
		return nil, err
	}

	allows := make(map[LogLevel]bool)
	for _, allowLevel := range allowList {
		allows[allowLevel] = true
	}

	return &filterDispatcher{disp, allows}, nil
}

func (filter *filterDispatcher) Dispatch(
	message string,
	level LogLevel,
	context logContextInterface,
	errorFunc func(err error)) {
	isAllowed, ok := filter.allowList[level]
	if ok && isAllowed {
		filter.dispatcher.Dispatch(message, level, context, errorFunc)
	}
}

func (filter *filterDispatcher) String() string {
	return fmt.Sprintf("filterDispatcher ->\n%s", filter.dispatcher)
}
