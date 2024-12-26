package btcwire

import (
	"bytes"
	"io"
)

const MaxTxInSequenceNum uint32 = 0xffffffff

type OutPoint struct {
	Hash  ShaHash
	Index uint32
}

func NewOutPoint(hash *ShaHash, index uint32) *OutPoint {
	return &OutPoint{
		Hash:  *hash,
		Index: index,
	}
}

type TxIn struct {
	PreviousOutpoint OutPoint
	SignatureScript  []byte
	Sequence         uint32
}

func NewTxIn(prevOut *OutPoint, signatureScript []byte) *TxIn {
	return &TxIn{
		PreviousOutpoint: *prevOut,
		SignatureScript:  signatureScript,
		Sequence:         MaxTxInSequenceNum,
	}
}

type TxOut struct {
	Value    int64
	PkScript []byte
}

func NewTxOut(value int64, pkScript []byte) *TxOut {
	return &TxOut{
		Value:    value,
		PkScript: pkScript,
	}
}

type MsgTx struct {
	Version  uint32
	TxIn     []*TxIn
	TxOut    []*TxOut
	LockTime uint32
}

func (msg *MsgTx) AddTxIn(ti *TxIn) {
	msg.TxIn = append(msg.TxIn, ti)
}

func (msg *MsgTx) AddTxOut(to *TxOut) {
	msg.TxOut = append(msg.TxOut, to)
}

func (tx *MsgTx) TxSha(pver uint32) (ShaHash, error) {
	var txsha ShaHash
	var wbuf bytes.Buffer
	err := tx.BtcEncode(&wbuf, pver)
	if err != nil {
		return txsha, err
	}
	txsha.SetBytes(DoubleSha256(wbuf.Bytes()))
	return txsha, nil
}

func (tx *MsgTx) Copy() *MsgTx {
	// Create new tx and start by copying primitive values.
	newTx := MsgTx{
		Version:  tx.Version,
		LockTime: tx.LockTime,
	}
	// Deep copy the old TxIn data.
	for _, oldTxIn := range tx.TxIn {
		// Deep copy the old previous outpoint.
		oldOutPoint := oldTxIn.PreviousOutpoint
		newOutPoint := OutPoint{}
		newOutPoint.Hash.SetBytes(oldOutPoint.Hash[:])
		newOutPoint.Index = oldOutPoint.Index
		// Deep copy the old signature script.
		var newScript []byte
		oldScript := oldTxIn.SignatureScript
		oldScriptLen := len(oldScript)
		if oldScriptLen > 0 {
			newScript = make([]byte, oldScriptLen, oldScriptLen)
			copy(newScript, oldScript[:oldScriptLen])
		}
		// Create new txIn with the deep copied data and append it to
		// new Tx.
		newTxIn := TxIn{
			PreviousOutpoint: newOutPoint,
			SignatureScript:  newScript,
			Sequence:         oldTxIn.Sequence,
		}
		newTx.TxIn = append(newTx.TxIn, &newTxIn)
	}
	// Deep copy the old TxOut data.
	for _, oldTxOut := range tx.TxOut {
		// Deep copy the old PkScript
		var newScript []byte
		oldScript := oldTxOut.PkScript
		oldScriptLen := len(oldScript)
		if oldScriptLen > 0 {
			newScript = make([]byte, oldScriptLen, oldScriptLen)
			copy(newScript, oldScript[:oldScriptLen])
		}
		// Create new txOut with the deep copied data and append it to
		// new Tx.
		newTxOut := TxOut{
			Value:    oldTxOut.Value,
			PkScript: newScript,
		}
		newTx.TxOut = append(newTx.TxOut, &newTxOut)
	}
	return &newTx
}

func (msg *MsgTx) BtcDecode(r io.Reader, pver uint32) error {
	err := readElement(r, &msg.Version)
	if err != nil {
		return err
	}
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}
	for i := uint64(0); i < count; i++ {
		ti := TxIn{}
		err = readTxIn(r, pver, msg.Version, &ti)
		if err != nil {
			return err
		}
		msg.TxIn = append(msg.TxIn, &ti)
	}
	count, err = readVarInt(r, pver)
	if err != nil {
		return err
	}
	for i := uint64(0); i < count; i++ {
		to := TxOut{}
		err = readTxOut(r, pver, msg.Version, &to)
		if err != nil {
			return err
		}
		msg.TxOut = append(msg.TxOut, &to)
	}
	err = readElement(r, &msg.LockTime)
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgTx) BtcEncode(w io.Writer, pver uint32) error {
	err := writeElement(w, msg.Version)
	if err != nil {
		return err
	}
	count := uint64(len(msg.TxIn))
	err = writeVarInt(w, pver, count)
	if err != nil {
		return err
	}
	for _, ti := range msg.TxIn {
		err = writeTxIn(w, pver, msg.Version, ti)
		if err != nil {
			return err
		}
	}
	count = uint64(len(msg.TxOut))
	err = writeVarInt(w, pver, count)
	if err != nil {
		return err
	}
	for _, to := range msg.TxOut {
		err = writeTxOut(w, pver, to)
		if err != nil {
			return err
		}
	}
	err = writeElement(w, msg.LockTime)
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgTx) Command() string {
	return cmdTx
}

func (msg *MsgTx) MaxPayloadLength(pver uint32) uint32 {
	return maxMessagePayload
}

func NewMsgTx() *MsgTx {
	return &MsgTx{Version: TxVersion}
}

func readOutPoint(r io.Reader, pver uint32, version uint32, op *OutPoint) error {
	err := readElements(r, &op.Hash, &op.Index)
	if err != nil {
		return err
	}
	return nil
}

func writeOutPoint(w io.Writer, pver uint32, version uint32, op *OutPoint) error {
	err := writeElements(w, op.Hash, op.Index)
	if err != nil {
		return err
	}
	return nil
}

func readTxIn(r io.Reader, pver uint32, version uint32, ti *TxIn) error {
	op := OutPoint{}
	err := readOutPoint(r, pver, version, &op)
	if err != nil {
		return err
	}
	ti.PreviousOutpoint = op
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}
	b := make([]byte, count)
	err = readElement(r, b)
	if err != nil {
		return err
	}
	ti.SignatureScript = b
	err = readElement(r, &ti.Sequence)
	if err != nil {
		return err
	}
	return nil
}

func writeTxIn(w io.Writer, pver uint32, version uint32, ti *TxIn) error {
	err := writeOutPoint(w, pver, version, &ti.PreviousOutpoint)
	if err != nil {
		return err
	}
	slen := uint64(len(ti.SignatureScript))
	err = writeVarInt(w, pver, slen)
	if err != nil {
		return err
	}
	b := []byte(ti.SignatureScript)
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	err = writeElement(w, &ti.Sequence)
	if err != nil {
		return err
	}
	return nil
}

// readTxOut reads the next sequence of bytes from r as a transaction output
// (TxOut).
func readTxOut(r io.Reader, pver uint32, version uint32, to *TxOut) error {
	err := readElement(r, &to.Value)
	if err != nil {
		return err
	}
	slen, err := readVarInt(r, pver)
	if err != nil {
		return err
	}
	b := make([]byte, slen)
	err = readElement(r, b)
	if err != nil {
		return err
	}
	to.PkScript = b
	return nil
}

// writeTxOut encodes to into the bitcoin protocol encoding for a transaction
// output (TxOut) to w.
func writeTxOut(w io.Writer, pver uint32, to *TxOut) error {
	err := writeElement(w, to.Value)
	if err != nil {
		return err
	}
	pkLen := uint64(len(to.PkScript))
	err = writeVarInt(w, pver, pkLen)
	if err != nil {
		return err
	}
	err = writeElement(w, to.PkScript)
	if err != nil {
		return err
	}
	return nil
}
