package metainfo

import (
	"bytes"
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

func TestParseMetainfoTwo(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "testData", "multitest.torrent"))
	if err != nil {
		t.Fatal("Failed to open torrent file: ", err)
	}

	m, err := ParseMetainfo(f)
	if err != nil {
		t.Fatal("Failed to parse metainfo file: ", err)
	}

	if m.Name != "multitest" {
		t.Error("Incorrect name: ", m.Name)
	}
	if len(m.AnnounceList) != 4 || m.AnnounceList[2] != "udp://tracker.istole.it:80" {
		t.Error("Incorrect announce list: ", m.AnnounceList)
	}
	if m.PieceCount != 6 {
		t.Error("Incorrect piece count: ", m.PieceCount)
	}
	if m.PieceLength != 16384 {
		t.Error("Incorrect piece length: ", m.PieceLength)
	}
	if len(m.Files) != 3 {
		t.Error("Incorrect number of files", m.Files)
	}
	if m.Files[0].Path != "multitest/test3.txt" || m.Files[0].Length != 36880 {
		t.Error("Incorrect file [1]: ", m.Files[0])
	}
	if m.Files[1].Path != "multitest/test2.txt" || m.Files[1].Length != 34113 {
		t.Error("Incorrect file [2]: ", m.Files[1])
	}
	if m.Files[2].Path != "multitest/test1.txt" || m.Files[2].Length != 24893 {
		t.Error("Incorrect file [3]: ", m.Files[2])
	}
	if !bytes.Equal(m.InfoHash[0:5], []byte{0x7f, 0x2e, 0x65, 0x2c, 0xda}) {
		t.Error("Incorrect infoshash: ", m.InfoHash)
	}
}
