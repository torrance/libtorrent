package libtorrent

import (
	"testing"
)

func TestAppendingAndRemovingSwarmTally(t *testing.T) {
	st := make(swarmTally, 14)

	bitf := []byte{0x81, 0x40}
	st.AddBitfield(bitf)

	if !equalInts(st, []int{1, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0}) {
		t.Error("Swarm tally incorrect, got [1]: ", st)
	}

	bitf = []byte{0x80, 0xC0}
	st.AddBitfield(bitf)
	if !equalInts(st, []int{2, 0, 0, 0, 0, 0, 0, 1, 1, 2, 0, 0, 0, 0}) {
		t.Error("Swarm tally incorrect, got [2]: ", st)
	}

	bitf = []byte{0x0, 0x40}
	st.RemoveBitfield(bitf)
	if !equalInts(st, []int{2, 0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0}) {
		t.Error("Swarm tally incorrect, got [3]: ", st)
	}
}
