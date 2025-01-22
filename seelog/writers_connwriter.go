package seelog

import (
	"fmt"
	"io"
	"net"
)

// connWriter is used to write to a stream-oriented network connection.
type connWriter struct {
	innerWriter    io.WriteCloser
	reconnectOnMsg bool
	recconect      bool
	net            string
	addr           string
}

// Creates writer to the address addr on the network netName.
// Connection will be opened on each write if reconnectOnMsg = true
func newConnWriter(netName string, addr string, reconnectOnMsg bool) *connWriter {
	newWriter := new(connWriter)

	newWriter.net = netName
	newWriter.addr = addr
	newWriter.reconnectOnMsg = reconnectOnMsg

	return newWriter
}

func (connWriter *connWriter) Close() error {
	if connWriter.innerWriter == nil {
		return nil
	}

	return connWriter.innerWriter.Close()
}

func (connWriter *connWriter) Write(bytes []byte) (n int, err error) {
	if connWriter.neddedConnectOnMsg() {
		err = connWriter.connect()
		if err != nil {
			return 0, err
		}
	}

	if connWriter.reconnectOnMsg {
		defer connWriter.innerWriter.Close()
	}

	n, err = connWriter.innerWriter.Write(bytes)
	if err != nil {
		connWriter.recconect = true
	}

	return
}

func (connWriter *connWriter) String() string {
	return fmt.Sprintf("Conn writer: [%s, %s, %v]", connWriter.net, connWriter.addr, connWriter.reconnectOnMsg)
}

func (connWriter *connWriter) connect() error {
	if connWriter.innerWriter != nil {
		connWriter.innerWriter.Close()
		connWriter.innerWriter = nil
	}

	conn, err := net.Dial(connWriter.net, connWriter.addr)
	if err != nil {
		return err
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if ok {
		tcpConn.SetKeepAlive(true)
	}

	connWriter.innerWriter = conn

	return nil
}

func (connWriter *connWriter) neddedConnectOnMsg() bool {
	if connWriter.recconect {
		connWriter.recconect = false
		return true
	}

	if connWriter.innerWriter == nil {
		return true
	}

	return connWriter.reconnectOnMsg
}
