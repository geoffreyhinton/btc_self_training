package btcwire

import (
	"bytes"
	"io"
	"time"
)

const BlockVersion uint32 = 2
const maxBlockHeaderPayload = 16 + maxVarIntPayload + (HashSize * 2)

type BlockHeader struct {
	Version    uint32
	PrevBlock  ShaHash
	MerkleRoot ShaHash
	Timestamp  time.Time
	Bits       uint32
	Nonce      uint32
	TxnCount   uint64
}

const blockHashLen = 80

func (h *BlockHeader) BlockSha(pver uint32) (ShaHash, error) {
	var buf bytes.Buffer
	err = writeBlockHeader(&buf, pver, h)
	if err != nil {
		return ShaHash{}, err
	}

	err = sha.SetBytes(DoubleSha256(buf.Bytes()[0:blockHashLen]))

	if err != nil {
		return ShaHash{}, err
	}
	return sha, nil
}

func NewBlockHeader(prevHash *ShaHash, merkleRootHash *ShaHash, bits uint32, nonce uint32) *BlockHeader {
	return &BlockHeader{
		Version:    BlockVersion,
		PrevBlock:  *prevHash,
		MerkleRoot: *merkleRootHash,
		Timestamp:  time.Now(),
		Bits:       bits,
		Nonce:      nonce,
		TxnCount:   0,
	}
}

func readBlockHeader(r io.Reader, pver uint32, bh *BlockHeader) error {
	var sec uint32

	err := readElements(r, &bh.Version, &bh.PrevBlock, &bh.MerkleRoot, &sec, &bh.Bits, &bh.Nonce)
	if err != nil {
		return err
	}

	bh.Timestamp = time.Unix(int64(sec), 0)
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}
	bh.TxnCount = count
	return nil
}

func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	sec := uint32(bh.Timestamp.Unix())
	err := writeElements(w, bh.Version, bh.PrevBlock, bh.MerkleRoot,
		sec, bh.Bits, bh.Nonce)
	if err != nil {
		return err
	}
	err = writeVarInt(w, pver, bh.TxnCount)
	if err != nil {
		return err
	}
	return nil
}
