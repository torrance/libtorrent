package libtorrent

import (
	"errors"
	"fmt"
)

type swarmTally []int

func (st swarmTally) AddBitfield(bitf *bitfield) (err error) {
	requiredLength := bitfieldLength(len(st))
	if requiredLength != len(bitf.field) {
		err = errors.New(fmt.Sprintf("addBitfield: Supplied bitfield incorrect size, want %d, got %d", requiredLength, len(bitf.field)))
		return
	}

	for i := 0; i < len(st); i++ {
		if st[i] == -1 {
			// We have this piece.
			continue
		}
		if bitf.Get(i) {
			st[i]++
		}
	}
	return
}

func (st swarmTally) RemoveBitfield(bitf *bitfield) (err error) {
	requiredLength := bitfieldLength(len(st))
	if requiredLength != len(bitf.field) {
		err = errors.New(fmt.Sprintf("removeBitfield: Supplied bitfield incorrect size, want %d, got %d", requiredLength, len(bitf.field)))
		return
	}

	for i := 0; i < len(st); i++ {
		if st[i] <= 0 {
			// We either have this piece, or something's gone wrong. Either way, leave as is.
			continue
		}
		if bitf.Get(i) {
			st[i]--
		}
	}
	return
}

func (st swarmTally) Zero() {
	for i := 0; i < len(st); i++ {
		if st[i] != -1 {
			st[i] = 0
		}
	}
}
