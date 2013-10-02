package libtorrent

import (
	"bytes"
	"testing"
)

func TestBitfieldSetTrue(t *testing.T) {
	bf := newBitfield(14)
	bf.SetTrue(0)
	bf.SetTrue(7)
	bf.SetTrue(9)
	if !bytes.Equal(bf, []byte{0x81, 0x40}) {
		t.Errorf("Bitfield SetTrue failed, got: %x", bf)
	}
}

func TestBitfieldGet(t *testing.T) {
	bf := newBitfield(14)
	bf.SetTrue(0)
	bf.SetTrue(7)
	bf.SetTrue(9)

	a := bf.Get(0)
	b := bf.Get(5)
	c := bf.Get(9)
	d := bf.Get(13)

	if !a || b || !c || d {
		t.Error("Bitfield Get failed")
	}
}
