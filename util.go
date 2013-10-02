package libtorrent

import (
	"encoding/binary"
	"io"
)

type monadWriter struct {
	w   io.Writer
	err error
}

func (mw *monadWriter) Write(i interface{}) {
	if mw.err != nil {
		return
	}
	mw.err = binary.Write(mw.w, binary.BigEndian, i)
}

type monadReader struct {
	r   io.Reader
	err error
}

func (mr *monadReader) Read(i interface{}) {
	if mr.err != nil {
		return
	}
	mr.err = binary.Read(mr.r, binary.BigEndian, i)
}

func equalInts(i []int, j []int) bool {
	if len(i) != len(j) {
		return false
	}
	for k := 0; k < len(i); k++ {
		if i[k] != j[k] {
			return false
		}
	}
	return true
}
