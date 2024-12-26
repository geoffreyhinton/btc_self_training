package btcwire

import (
	"fmt"
	"io"
)

const MaxAddrPerMsg = 1000

type MsgAddr struct {
	AddrList []*NetAddress
}

func (msg *MsgAddr) AddAddress(na *NetAddress) error {
	if len(msg.AddrList)+1 > MaxAddrPerMsg {
		str := "MsgAddr.AddAddress: too many addresses for message [max %v]"
		return fmt.Errorf(str, MaxAddrPerMsg)
	}
	msg.AddrList = append(msg.AddrList, na)
	return nil
}

func (msg *MsgAddr) AddAddresses(netAddrs ...*NetAddress) error {
	for _, na := range netAddrs {
		err := msg.AddAddress(na)
		if err != nil {
			return err
		}
	}
	return nil
}

func (msg *MsgAddr) ClearAddresses() {
	msg.AddrList = []*NetAddress{}
}

func (msg *MsgAddr) BtcDecode(r io.Reader, pver uint32) error {
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}
	// Limit to max addresses per message.
	if count > MaxAddrPerMsg {
		str := "MsgAddr.BtcDecode: too many addresses in message [%v]"
		return fmt.Errorf(str, count)
	}
	for i := uint64(0); i < count; i++ {
		na := NetAddress{}
		err := readNetAddress(r, pver, &na, true)
		if err != nil {
			return err
		}
		msg.AddAddress(&na)
	}
	return nil
}
func (msg *MsgAddr) BtcEncode(w io.Writer, pver uint32) error {
	// Protocol versions before MultipleAddressVersion only allowed 1 address
	// per message.
	count := len(msg.AddrList)
	if pver < MultipleAddressVersion && count > 1 {
		str := "MsgAddr.BtcDecode: too many addresses in message " +
			"for protocol version [version %v max 1]"
		return fmt.Errorf(str, pver)
	}
	if count > MaxAddrPerMsg {
		str := "MsgAddr.BtcDecode: too many addresses in message [max %v]"
		return fmt.Errorf(str, count)
	}
	err := writeVarInt(w, pver, uint64(count))
	if err != nil {
		return err
	}
	for _, na := range msg.AddrList {
		err = writeNetAddress(w, pver, na, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgAddr) Command() string {
	return cmdAddr
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgAddr) MaxPayloadLength(pver uint32) uint32 {
	if pver < MultipleAddressVersion {
		// Num addresses (varInt) + a single net addresses.
		return maxVarIntPayload + maxNetAddressPayload(pver)
	}
	// Num addresses (varInt) + max allowed addresses.
	return maxVarIntPayload + (MaxAddrPerMsg * maxNetAddressPayload(pver))
}

// NewMsgAddr returns a new bitcoin addr message that conforms to the
// Message interface.  See MsgAddr for details.
func NewMsgAddr() *MsgAddr {
	return &MsgAddr{}
}
