package seelog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// fileWriter is used to write to a file.
type fileWriter struct {
	innerWriter io.WriteCloser
	fileName    string
}

// Creates a new file and a corresponding writer. Returns error, if the file couldn't be created.
func newFileWriter(fileName string) (writer *fileWriter, err error) {
	newWriter := new(fileWriter)
	newWriter.fileName = fileName

	return newWriter, nil
}

func (fw *fileWriter) Close() error {
	if fw.innerWriter != nil {
		return fw.innerWriter.Close()
	}
	return nil
}

// Create folder and file on WriteLog/Write first call
func (fw *fileWriter) Write(bytes []byte) (n int, err error) {
	if fw.innerWriter == nil {
		if err := fw.createFile(); err != nil {
			return 0, err
		}
	}
	return fw.innerWriter.Write(bytes)
}

func (fw *fileWriter) createFile() error {
	folder, _ := filepath.Split(fw.fileName)
	var err error

	if 0 != len(folder) {
		err = os.MkdirAll(folder, defaultDirectoryPermissions)
		if err != nil {
			return err
		}
	}

	// If exists
	_, err = os.Lstat(fw.fileName)
	if nil == err {
		fw.innerWriter, err = os.OpenFile(fw.fileName, os.O_WRONLY|os.O_APPEND, defaultFilePermissions)
	} else {
		fw.innerWriter, err = os.Create(fw.fileName)
	}

	if err != nil {
		return err
	}

	return nil
}

func (fw *fileWriter) String() string {
	return fmt.Sprintf("File writer: %s", fw.fileName)
}
