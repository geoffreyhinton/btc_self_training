package btcwire

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math"
)

const maxVarIntPayload = 9

func readElement(r io.Reader, element interface{}) error {
	return binary.Read(r, binary.LittleEndian, element)
}

func readElements(r io.Reader, elements ...interface{}) error {
	for _, element := range elements {
		if err := readElement(r, element); err != nil {
			return err
		}
	}
	return nil
}

func writeElement(w io.Writer, element interface{}) error {
	return binary.Write(w, binary.LittleEndian, element)
}

func writeElements(w io.Writer, elements ...interface{}) error {
	for _, element := range elements {
		if err := writeElement(w, element); err != nil {
			return err
		}
	}
	return nil
}

func readVarInt(r io.Reader, pver uint32) (uint64, error) {
	b := make([]byte, 1)
	_, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	var rv uint64
	discriminant := uint8(b[0])
	switch discriminant {
	case 0xff:
		var u uint64
		err = binary.Read(r, binary.LittleEndian, &u)
		if err != nil {
			return 0, err
		}
		rv = u
	case 0xfe:
		var u uint32
		err = binary.Read(r, binary.LittleEndian, &u)
		if err != nil {
			return 0, err
		}
		rv = uint64(u)
	case 0xfd:
		var u uint16
		err = binary.Read(r, binary.LittleEndian, &u)
		if err != nil {
			return 0, err
		}
		rv = uint64(u)
	default:
		rv = uint64(discriminant)
	}
	return rv, nil
}

func writeVarInt(w io.Writer, pver uint32, val uint64) error {
	if val > math.MaxUint32 {
		err := writeElements(w, []byte{0xff}, uint64(val))
		if err != nil {
			return err
		}
		return nil
	}
	if val > math.MaxUint16 {
		err := writeElements(w, []byte{0xfe}, uint32(val))
		if err != nil {
			return err
		}
		return nil
	}
	if val >= 0xfd {
		err := writeElements(w, []byte{0xfd}, uint16(val))
		if err != nil {
			return err
		}
		return nil
	}
	return writeElement(w, uint8(val))
}

func readVarString(r io.Reader, pver uint32) (string, error) {
	slen, err := readVarInt(r, pver)
	if err != nil {
		return "", err
	}
	buf := make([]byte, slen)
	err = readElement(r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func writeVarString(w io.Writer, pver uint32, str string) error {
	err := writeVarInt(w, pver, uint64(len(str)))
	if err != nil {
		return err
	}
	err = writeElement(w, []byte(str))
	if err != nil {
		return err
	}
	return nil
}

func randomUint64(r io.Reader) (uint64, error) {
	b := make([]byte, 8)
	n, err := r.Read(b)
	if n != len(b) {
		return 0, io.ErrShortBuffer
	}
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

func RandomUint64() (uint64, error) {
	return randomUint64(rand.Reader)
}

func DoubleSha256(b []byte) []byte {
	hasher := sha256.New()
	hasher.Write(b)
	sum := hasher.Sum(nil)
	hasher.Reset()
	hasher.Write(sum)
	sum = hasher.Sum(nil)
	return sum
}
