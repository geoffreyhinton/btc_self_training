package seelog

import "fmt"

// consoleWriter is used to write to console
type consoleWriter struct {
}

// Creates a new console writer. Returns error, if the console writer couldn't be created.
func newConsoleWriter() (writer *consoleWriter, err error) {
	newWriter := new(consoleWriter)

	return newWriter, nil
}

// Create folder and file on WriteLog/Write first call
func (console *consoleWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}

func (console *consoleWriter) String() string {
	return "Console writer"
}
