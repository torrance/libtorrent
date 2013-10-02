package libtorrent

import (
	"errors"
	"io"
	"io/ioutil"
)

type bitfield []uint8

func newBitfield(length int) (bf bitfield) {
	bf = make([]byte, bitfieldLength(length))
	return
}

func parseBitfield(r io.Reader) (bf bitfield, err error) {
	bf, err = ioutil.ReadAll(r)
	return
}

func (bf bitfield) SetTrue(index int) (err error) {
	if index >= len(bf)*8 {
		err = errors.New("Bitfield error: Index out of range")
	}

	byteIndex := index / 8
	bitOffset := index % 8

	bitMask := 1 << (7 - uint8(bitOffset))
	bf[byteIndex] |= uint8(bitMask)
	return
}

func (bf bitfield) Get(index int) (b bool) {
	if index >= len(bf)*8 {
		b = false
		return
	}

	byteIndex := index / 8
	bitOffset := index % 8

	bitMask := uint8(1 << (7 - uint8(bitOffset)))
	b = bitMask == bf[byteIndex]&bitMask
	return
}

func (bf bitfield) BinaryDump(w io.Writer) (err error) {
	_, err = w.Write(bf)
	return
}

func bitfieldLength(i int) (length int) {
	// Check that supplied bitfield is correct length
	length = i / 8
	if i%8 != 0 {
		length++
	}
	return
}
