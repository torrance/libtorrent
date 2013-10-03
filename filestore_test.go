package libtorrent

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetBlockWithSingleFile(t *testing.T) {

	file1 := testTorrentStorer{
		reader: bytes.NewReader([]byte{1, 2, 3, 4}),
	}
	hashes := [][]byte{[]byte{1}, []byte{2}}
	fs, err := newFileStore([]torrentStorer{file1}, hashes, 3)
	if err != nil {
		t.Fatalf("Failed to create filestore: %s", err)
	}

	block, err := fs.getBlock(0, 1, 2)
	if err != nil {
		t.Fatalf("Failed to get block [1]: %s", err)
	}

	if !bytes.Equal(block, []byte{2, 3}) {
		t.Errorf("Block contained incorrect values, got [1]: %x", block)
	}

	pieceLength := fs.getPieceLength(1)
	t.Logf("Piece length: %d", pieceLength)
	if block, err = fs.getBlock(1, 0, pieceLength); err != nil {
		t.Fatalf("Failed to get block [2]: %s", err)
	}

	if !bytes.Equal(block, []byte{4}) {
		t.Errorf("Block contained incorrect values, got [1]: %x", block)
	}
}

func TestGetBlockWithMultipleFiles(t *testing.T) {
	file1 := testTorrentStorer{
		reader: bytes.NewReader([]byte{1, 2, 3, 4}),
	}
	file2 := testTorrentStorer{
		reader: bytes.NewReader([]byte{5, 6, 7}),
	}
	file3 := testTorrentStorer{
		reader: bytes.NewReader([]byte{8, 9, 10, 11, 12, 13}),
	}

	b := []byte{1}
	hashes := [][]byte{b, b, b, b, b}
	fs, err := newFileStore([]torrentStorer{file1, file2, file3}, hashes, 3)
	if err != nil {
		t.Fatalf("Failed to create filestore: %s", err)
	}

	// Test 1: only select from first file
	block, err := fs.getBlock(0, 1, 2)
	if err != nil {
		t.Fatalf("Failed to get block [1]: %s", err)
	}
	if !bytes.Equal(block, []byte{2, 3}) {
		t.Errorf("Block contained incorrect values, got [1]: %x", block)
	}

	// Test 2: select from second file only
	if block, err = fs.getBlock(1, 1, 2); err != nil {
		t.Fatalf("Failed to get block [2]: %s", err)
	}
	if !bytes.Equal(block, []byte{5, 6}) {
		t.Errorf("Block contained incorrect values, got [2]: %x", block)
	}

	// Test 3: select from piece bridging two files
	if block, err = fs.getBlock(2, 0, 3); err != nil {
		t.Fatalf("Failed to get block [3]: %s", err)
	}
	if !bytes.Equal(block, []byte{7, 8, 9}) {
		t.Errorf("Block contained incorrect values, got [3]: %x", block)
	}

	// Test 4: select last piece
	if block, err = fs.getBlock(4, 0, fs.getPieceLength(4)); err != nil {
		t.Fatalf("Failed to get block [4]: %s", err)
	}
	if !bytes.Equal(block, []byte{13}) {
		t.Errorf("Block contained incorrect values, got [4]: %x", block)
	}
}

type testTorrentStorer struct {
	reader *bytes.Reader
}

func (stor testTorrentStorer) ReadAt(b []byte, off int64) (n int, err error) {
	return stor.reader.ReadAt(b, off)
}

func (stor testTorrentStorer) length() int64 {
	return int64(stor.reader.Len())
}

func TestNewTFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "libtorrentTesting")
	if err != nil {
		t.Fatal("Could not create temporary directory to run tests: ", err)
	}
	defer os.RemoveAll(tmpDir)

	tfile, err := newTorrentFile(tmpDir, filepath.Join("dir1", "dir2", "file.txt"), 1234)
	if err != nil {
		t.Fatal(err)
	}

	if tfile.length() != 1234 {
		t.Error("TFile.length not correctly set. Actual value: ", tfile.length)
	}
	if tfile.path != filepath.Join("dir1", "dir2", "file.txt") {
		t.Error("TFile.filename not correctly set. Actual value: ", tfile.path)
	}
}

func TestGetBlockWithRealFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "libtorrentTesting")
	if err != nil {
		t.Fatal("Could not create temporary directory to run tests: ", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile, _ := os.Create(tmpDir + string(os.PathSeparator) + "test.txt")
	originalFile, _ := os.Open("testData/test.txt")
	io.Copy(testFile, originalFile)

	tfile, err := newTorrentFile(tmpDir, "test.txt", 36880)
	if err != nil {
		t.Fatal("Could not create tfile: ", err)
	}

	hashes := [][]byte{
		[]byte{235, 203, 167, 177, 114, 56, 226, 172, 6, 254, 96, 77, 68, 107, 80, 141, 148, 248, 180, 189},
		[]byte{199, 10, 10, 118, 99, 244, 176, 96, 247, 53, 217, 230, 10, 42, 50, 233, 147, 116, 217, 141},
	}

	fs, err := newFileStore([]torrentStorer{tfile}, hashes, 32768)
	if err != nil {
		t.Fatal("Couldn't create fileStore: ", err)
	}

	block, err := fs.getBlock(0, 5, 5)
	if err != nil {
		t.Error("Error calling getBlock: ", err)
	}
	if !bytes.Equal(block, []byte{0x69, 0x73, 0x20, 0x69, 0x73}) {
		t.Errorf("Loaded incorrect block data [1]: %x", block)
	}

	block, err = fs.getBlock(1, 5, 5)
	if err != nil {
		t.Error("Error calling getBlock: ", err)
	}
	if !bytes.Equal(block, []byte{0x6e, 0x74, 0x2e, 0x0a, 0x32}) {
		t.Errorf("Loaded incorrect block data [2]: %x", block)
	}

	// Test loading final piece
	block, err = fs.getBlock(1, 0, fs.getPieceLength(1))
	if err != nil {
		t.Error("Error calling getBlock: ", err)
	}
	if !bytes.Equal(block[:3], []byte{0x70, 0x6f, 0x72}) || !bytes.Equal(block[len(block)-3:], []byte{0x74, 0x2e, 0x0a}) {
		t.Errorf("Loaded incorrect block data [3]: Start: %x End: %x\n", block[:3], block[len(block)-3:])
	}

	bitf, err := fs.validate()
	if err != nil {
		t.Error("Error calling validate: ", err)
	}

	if len(bitf.field) != 1 || bitf.field[0] != 0xc0 {
		t.Errorf("Incorrect bitfield, got: %x", bitf.field)
	}

}
