package libtorrent

import (
	"github.com/torrance/libtorrent/bitfield"
	"testing"
)

func TestAppendingAndRemovingSwarmTally(t *testing.T) {
	st := make(swarmTally, 14)

	bitf := bitfield.NewBitfield(14)
	bitf.SetTrue(0)
	bitf.SetTrue(7)
	bitf.SetTrue(9)
	st.AddBitfield(bitf)

	if !equalInts(st, []int{1, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0}) {
		t.Error("Swarm tally incorrect, got [1]: ", st)
	}

	bitf = bitfield.NewBitfield(14)
	bitf.SetTrue(0)
	bitf.SetTrue(8)
	bitf.SetTrue(9)
	st.AddBitfield(bitf)
	if !equalInts(st, []int{2, 0, 0, 0, 0, 0, 0, 1, 1, 2, 0, 0, 0, 0}) {
		t.Error("Swarm tally incorrect, got [2]: ", st)
	}

	bitf = bitfield.NewBitfield(14)
	bitf.SetTrue(9)
	st.RemoveBitfield(bitf)
	if !equalInts(st, []int{2, 0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0}) {
		t.Error("Swarm tally incorrect, got [3]: ", st)
	}
}
