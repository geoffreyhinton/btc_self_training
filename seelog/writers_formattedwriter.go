package seelog

import (
	"errors"
	"fmt"
	"io"
)

type formattedWriter struct {
	writer    io.Writer
	formatter *formatter
}

func newFormattedWriter(writer io.Writer, formatter *formatter) (*formattedWriter, error) {
	if formatter == nil {
		return nil, errors.New("formatter can not be nil")
	}

	return &formattedWriter{writer, formatter}, nil
}

func (formattedWriter *formattedWriter) Write(message string, level LogLevel, context logContextInterface) error {
	str := formattedWriter.formatter.Format(message, level, context)
	_, err := formattedWriter.writer.Write([]byte(str))
	return err
}

func (formattedWriter *formattedWriter) String() string {
	return fmt.Sprintf("writer: %s, format: %s", formattedWriter.writer, formattedWriter.formatter)
}

func (formattedWriter *formattedWriter) Writer() io.Writer {
	return formattedWriter.writer
}

func (formattedWriter *formattedWriter) Format() *formatter {
	return formattedWriter.formatter
}
