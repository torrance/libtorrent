package libtorrent

import (
	"errors"
	"io"
	"io/ioutil"
)

type bitfield struct {
	length int
	field  []uint8
}

func newBitfield(length int) (bf *bitfield) {
	bf = &bitfield{
		length: length,
		field:  make([]byte, bitfieldLength(length)),
	}
	return
}

func parseBitfield(r io.Reader) (bf *bitfield, err error) {
	field, err := ioutil.ReadAll(r)
	bf = &bitfield{
		field: field,
	}
	return
}

func (bf *bitfield) SetLength(length int) error {
	if length > len(bf.field)*8 {
		return errors.New("Attempted to set bitfield length larger than underlying bitfield")
	}
	bf.length = length
	return nil
}

func (bf *bitfield) SetTrue(index int) (err error) {
	if (bf.length > 0 && index >= bf.length) || (bf.length == 0 && index >= len(bf.field)*8) {
		err = errors.New("Bitfield error: Index out of range")
	}
	bf.field[index>>3] |= 1 << (7 - uint(index)&7)
	return
}

func (bf *bitfield) Get(index int) bool {
	if (bf.length > 0 && index >= bf.length) || (bf.length == 0 && index >= len(bf.field)*8) {
		return false
	}

	return bf.field[index>>3]&(1<<(7-uint(index)&7)) != 0
}

func bitfieldLength(i int) (length int) {
	// Check that supplied bitfield is correct length
	length = i / 8
	if i%8 != 0 {
		length++
	}
	return
}
