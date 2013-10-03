package filestore

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
	fs, err := NewFileStore([]TorrentStorer{file1}, hashes, 3)
	if err != nil {
		t.Fatalf("Failed to create filestore: %s", err)
	}

	block, err := fs.GetBlock(0, 1, 2)
	if err != nil {
		t.Fatalf("Failed to get block [1]: %s", err)
	}

	if !bytes.Equal(block, []byte{2, 3}) {
		t.Errorf("Block contained incorrect values, got [1]: %x", block)
	}

	pieceLength := fs.getPieceLength(1)
	t.Logf("Piece length: %d", pieceLength)
	if block, err = fs.GetBlock(1, 0, pieceLength); err != nil {
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
	fs, err := NewFileStore([]TorrentStorer{file1, file2, file3}, hashes, 3)
	if err != nil {
		t.Fatalf("Failed to create filestore: %s", err)
	}

	// Test 1: only select from first file
	block, err := fs.GetBlock(0, 1, 2)
	if err != nil {
		t.Fatalf("Failed to get block [1]: %s", err)
	}
	if !bytes.Equal(block, []byte{2, 3}) {
		t.Errorf("Block contained incorrect values, got [1]: %x", block)
	}

	// Test 2: select from second file only
	if block, err = fs.GetBlock(1, 1, 2); err != nil {
		t.Fatalf("Failed to get block [2]: %s", err)
	}
	if !bytes.Equal(block, []byte{5, 6}) {
		t.Errorf("Block contained incorrect values, got [2]: %x", block)
	}

	// Test 3: select from piece bridging two files
	if block, err = fs.GetBlock(2, 0, 3); err != nil {
		t.Fatalf("Failed to get block [3]: %s", err)
	}
	if !bytes.Equal(block, []byte{7, 8, 9}) {
		t.Errorf("Block contained incorrect values, got [3]: %x", block)
	}

	// Test 4: select last piece
	if block, err = fs.GetBlock(4, 0, fs.getPieceLength(4)); err != nil {
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

func (stor testTorrentStorer) Length() int64 {
	return int64(stor.reader.Len())
}

func TestNewTFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "libtorrentTesting")
	if err != nil {
		t.Fatal("Could not create temporary directory to run tests: ", err)
	}
	defer os.RemoveAll(tmpDir)

	tfile, err := NewTorrentFile(tmpDir, filepath.Join("dir1", "dir2", "file.txt"), 1234)
	if err != nil {
		t.Fatal(err)
	}

	if tfile.Length() != 1234 {
		t.Error("TFile.length not correctly set. Actual value: ", tfile.Length())
	}
	if tfile.path != filepath.Join("dir1", "dir2", "file.txt") {
		t.Error("TFile.filename not correctly set. Actual value: ", tfile.path)
	}
	fi, err := os.Stat(filepath.Join(tmpDir, "dir1", "dir2", "file.txt"))
	if err != nil {
		t.Fatal("Failed to stat file: ", err)
	}
	if fi.Size() != 1234 {
		t.Error("File not 1234bytes, actual value: ", fi.Size())
	}
}

func TestGetBlockWithRealFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "libtorrentTesting")
	if err != nil {
		t.Fatal("Could not create temporary directory to run tests: ", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile, _ := os.Create(filepath.Join(tmpDir, "test.txt"))
	originalFile, _ := os.Open(filepath.Join("..", "testData", "test.txt"))
	io.Copy(testFile, originalFile)

	tfile, err := NewTorrentFile(tmpDir, "test.txt", 36880)
	if err != nil {
		t.Fatal("Could not create tfile: ", err)
	}

	hashes := [][]byte{
		[]byte{235, 203, 167, 177, 114, 56, 226, 172, 6, 254, 96, 77, 68, 107, 80, 141, 148, 248, 180, 189},
		[]byte{199, 10, 10, 118, 99, 244, 176, 96, 247, 53, 217, 230, 10, 42, 50, 233, 147, 116, 217, 141},
	}

	fs, err := NewFileStore([]TorrentStorer{tfile}, hashes, 32768)
	if err != nil {
		t.Fatal("Couldn't create fileStore: ", err)
	}

	block, err := fs.GetBlock(0, 5, 5)
	if err != nil {
		t.Error("Error calling getBlock: ", err)
	}
	if !bytes.Equal(block, []byte{0x69, 0x73, 0x20, 0x69, 0x73}) {
		t.Errorf("Loaded incorrect block data [1]: %x", block)
	}

	block, err = fs.GetBlock(1, 5, 5)
	if err != nil {
		t.Error("Error calling getBlock: ", err)
	}
	if !bytes.Equal(block, []byte{0x6e, 0x74, 0x2e, 0x0a, 0x32}) {
		t.Errorf("Loaded incorrect block data [2]: %x", block)
	}

	// Test loading final piece
	block, err = fs.GetBlock(1, 0, fs.getPieceLength(1))
	if err != nil {
		t.Error("Error calling getBlock: ", err)
	}
	if !bytes.Equal(block[:3], []byte{0x70, 0x6f, 0x72}) || !bytes.Equal(block[len(block)-3:], []byte{0x74, 0x2e, 0x0a}) {
		t.Errorf("Loaded incorrect block data [3]: Start: %x End: %x\n", block[:3], block[len(block)-3:])
	}

	bitf, err := fs.Validate()
	if err != nil {
		t.Error("Error calling validate: ", err)
	}

	if bitf.ByteLength() != 1 || bitf.Bytes()[0] != 0xc0 {
		t.Errorf("Incorrect bitfield, got: %x", bitf.Bytes())
	}

}

func TestFileStoreWithMultipleFiles(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "libtorrentTesting")
	if err != nil {
		t.Fatal("Could not create temporary directory to run tests: ", err)
	}
	defer os.RemoveAll(tmpDir)

	// First create two empty TFiles. We'll test that they're created correctly.

	file1, err := NewTorrentFile(tmpDir, filepath.Join("multitest", "test1.txt"), 24893)
	if err != nil {
		t.Fatal("Failed to created file1: ", err)
	} else {
		if fi, err := file1.fd.Stat(); err != nil {
			t.Fatal("Failed to stat file 1: ", err)
		} else if fi.Size() != 24893 {
			t.Fatal("File 1 size incorrect, got: ", fi.Size())
		}
	}

	file2, err := NewTorrentFile(tmpDir, filepath.Join("multitest", "test2.txt"), 34113)
	if err != nil {
		t.Fatal("Failed to created file2: ", err)
	} else {
		if fi, err := file2.fd.Stat(); err != nil {
			t.Fatal("Failed to stat file 2: ", err)
		} else if fi.Size() != 34113 {
			t.Fatal("File 2 size incorrect, got: ", fi.Size())
		}
	}

	// Now replace these files with the actual files
	for _, fileName := range []string{"test1.txt", "test2.txt", "test3.txt"} {
		testFile, err := os.Create(filepath.Join(tmpDir, "multitest", fileName))
		if err != nil {
			t.Fatal("Failed to open test file: ", fileName, err)
		}
		originalFile, err := os.Open(filepath.Join("..", "testData", "multitest", fileName))
		if err != nil {
			t.Fatal("Failed to open original file: ", fileName, err)
		}
		_, err = io.Copy(testFile, originalFile)
		if err != nil {
			t.Fatal("Failed to copy files: ", fileName, err)
		}
	}

	// We add file 3 after having copied across the actual data, to ensure that
	// it doesn't delete any data.
	file3, err := NewTorrentFile(tmpDir, filepath.Join("multitest", "test3.txt"), 36880)
	if err != nil {
		t.Fatal("Failed to created file3: ", err)
	} else {
		if fi, err := file3.fd.Stat(); err != nil {
			t.Fatal("Failed to stat file 3: ", err)
		} else if fi.Size() != 36880 {
			t.Fatal("File 3 size incorrect, got: ", fi.Size())
		}
	}

	// Create FileStore object
	hashes := [][]byte{
		[]byte{73, 34, 176, 29, 229, 125, 157, 28, 41, 61, 161, 34, 149, 47, 162, 50, 32, 142, 179, 113},
		[]byte{253, 88, 132, 30, 179, 131, 129, 178, 163, 242, 219, 174, 160, 79, 92, 251, 81, 103, 81, 153},
		[]byte{103, 128, 108, 125, 224, 90, 210, 85, 56, 27, 112, 170, 148, 114, 155, 39, 132, 132, 67, 148},
		[]byte{207, 35, 173, 86, 79, 73, 78, 120, 174, 37, 56, 240, 209, 56, 179, 35, 216, 183, 95, 250},
		[]byte{234, 216, 33, 126, 28, 215, 199, 57, 176, 132, 212, 140, 74, 106, 140, 205, 81, 39, 183, 54},
		[]byte{81, 45, 79, 164, 142, 254, 170, 73, 121, 139, 104, 80, 249, 172, 251, 139, 232, 73, 69, 143},
	}

	fs, err := NewFileStore([]TorrentStorer{file3, file2, file1}, hashes, 16384)
	if err != nil {
		t.Fatal("Error creating filestore: ", err)
	}

	bitf, err := fs.Validate()
	if err != nil {
		t.Error("Error calling validate: ", err)
	}

	if bitf.ByteLength() != 1 || bitf.Bytes()[0] != 0xFC {
		t.Errorf("Incorrect bitfield, got: %x", bitf.Bytes())
	}
}
