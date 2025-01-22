package seelog

import (
	"fmt"
)

// A splitDispatcher just writes the given message to underlying receivers. (Splits the message stream.)
type splitDispatcher struct {
	*dispatcher
}

func newSplitDispatcher(formatter *formatter, receivers []interface{}) (*splitDispatcher, error) {
	disp, err := createDispatcher(formatter, receivers)
	if err != nil {
		return nil, err
	}

	return &splitDispatcher{disp}, nil
}

func (splitter *splitDispatcher) String() string {
	return fmt.Sprintf("splitDispatcher ->\n%s", splitter.dispatcher.String())
}
