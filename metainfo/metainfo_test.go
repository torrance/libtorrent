package metainfo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParseMetainfoOne(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "testData", "test.txt.torrent"))
	if err != nil {
		t.Fatal("Failed to open torrent file: ", err)
	}

	m, err := ParseMetainfo(f)
	if err != nil {
		t.Fatal("Failed to parse metainfo file: ", err)
	}

	if m.Name != "test.txt" {
		t.Error("Incorrect name: ", m.Name)
	}
	if len(m.AnnounceList) != 1 || m.AnnounceList[0] != "udp://tracker.openbittorrent.com:80/announce" {
		t.Error("Incorrect announce list: ", m.AnnounceList)
	}
	if m.PieceCount != 2 {
		t.Error("Incorrect piece count: ", m.PieceCount)
	}
	if m.PieceLength != 32768 {
		t.Error("Incorrect piece length: ", m.PieceLength)
	}
	if len(m.Files) != 1 || m.Files[0].Length != 36880 || m.Files[0].Path != "test.txt" {
		t.Error("Incorrect file data: ", m.Files)
	}
	if !bytes.Equal(m.InfoHash[0:5], []byte{116, 45, 71, 83, 15}) {
		t.Error("Incorrect infoshash: ", m.InfoHash)
	}
}
