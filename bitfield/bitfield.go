package bitfield

import (
	"errors"
	"io"
	"io/ioutil"
)

type Bitfield struct {
	length int
	field  []byte
}

func NewBitfield(length int) (bf *Bitfield) {
	bf = &Bitfield{
		length: length,
		field:  make([]byte, bitfieldLength(length)),
	}
	return
}

func ParseBitfield(r io.Reader) (bf *Bitfield, err error) {
	field, err := ioutil.ReadAll(r)
	bf = &Bitfield{
		field: field,
	}
	return
}

func (bf *Bitfield) Length() int {
	return bf.length
}

func (bf *Bitfield) SetLength(length int) error {
	if length > len(bf.field)*8 {
		return errors.New("Attempted to set bitfield length larger than underlying bitfield")
	}
	bf.length = length
	return nil
}

func (bf *Bitfield) SetTrue(index int) (err error) {
	if (bf.length > 0 && index >= bf.length) || (bf.length == 0 && index >= len(bf.field)*8) {
		err = errors.New("Bitfield error: Index out of range")
	}
	bf.field[index>>3] |= 1 << (7 - uint(index)&7)
	return
}

func (bf *Bitfield) Get(index int) bool {
	if (bf.length > 0 && index >= bf.length) || (bf.length == 0 && index >= len(bf.field)*8) {
		return false
	}

	return bf.field[index>>3]&(1<<(7-uint(index)&7)) != 0
}

func (bf *Bitfield) Bytes() []byte {
	return bf.field
}

func (bf *Bitfield) ByteLength() int {
	return len(bf.field)
}

func bitfieldLength(i int) (length int) {
	// Check that supplied bitfield is correct length
	length = i / 8
	if i%8 != 0 {
		length++
	}
	return
}
