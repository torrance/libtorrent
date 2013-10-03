package libtorrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type fileStore struct {
	tfiles      []torrentStorer
	hashes      [][]byte
	pieceLength int64
	totalLength int64
}

func newFileStore(tfiles []torrentStorer, hashes [][]byte, pieceLength int64) (fs *fileStore, err error) {
	fs = &fileStore{
		tfiles:      tfiles,
		hashes:      hashes,
		pieceLength: pieceLength,
	}

	for _, tfile := range tfiles {
		fs.totalLength += tfile.length()
	}

	return
}

func (fs *fileStore) validate() (bitf *bitfield, err error) {
	bitf = newBitfield(len(fs.hashes))

	for i, _ := range fs.hashes {
		var ok bool
		ok, err = fs.validatePiece(i)
		if err != nil {
			return
		} else if ok {
			bitf.SetTrue(i)
		}
	}
	return
}

func (fs *fileStore) validatePiece(index int) (ok bool, err error) {
	block, err := fs.getBlock(index, 0, fs.getPieceLength(index))
	if err != nil {
		return
	}

	h := sha1.New()
	h.Write(block)
	if bytes.Equal(h.Sum(nil), fs.hashes[index]) {
		ok = true
	}
	return
}

func (fs *fileStore) getPieceLength(index int) int64 {
	if index == len(fs.hashes)-1 {
		return fs.totalLength % fs.pieceLength
	} else {
		return fs.pieceLength
	}
}

func (fs *fileStore) getBlock(pieceIndex int, offset int64, length int64) (block []byte, err error) {
	if length+offset > fs.getPieceLength(pieceIndex) {
		err = errors.New("Requested block overran piece length")
		return
	}

	block = make([]byte, length)
	segment := block

	offset = int64(pieceIndex)*fs.pieceLength + offset

	for _, tfile := range fs.tfiles {
		var lengthRead int
		lengthRead, err = tfile.ReadAt(segment, offset)

		if err == nil {
			// We've read it all!
			break
		} else if err == io.EOF {
			// We haven't read anything, or only a partial read
			segment = segment[lengthRead:]
			if offset-tfile.length() < 0 {
				offset = 0
			} else {
				offset -= tfile.length()
			}
		} else if err != nil {
			// Something else went wrong
			break
		}
	}

	if err != nil {
		logger.Error("Failed to get block %d, %d, %d", pieceIndex, offset, length)
		logger.Error("Piece length: %d, Total Length: %d, Pieces: %d", fs.pieceLength, fs.totalLength, len(fs.hashes))
	}

	return
}

type torrentStorer interface {
	io.ReaderAt
	length() int64
}

type torrentFile struct {
	lth  int64
	path string
	fd   *os.File
}

func newTorrentFile(rootDirectory string, path string, length int64) (tfile *torrentFile, err error) {
	if len(path) == 0 {
		err = errors.New("Path must have at least 1 component.")
		return
	}

	// Root directory must already exist
	rootDirectoryFileInfo, err := os.Stat(rootDirectory)
	if err != nil {
		return
	}
	if !rootDirectoryFileInfo.IsDir() {
		err = errors.New(rootDirectory + " is not a directory")
		return
	}

	absPath := filepath.Join(rootDirectory, path)

	// Create any required parent directories
	dirs := filepath.Dir(absPath)
	if err = os.MkdirAll(dirs, 0755); err != nil {
		return
	}

	// Create or open file
	fd, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE, 0644)

	// Stat for size of file
	stat, err := fd.Stat()
	if err != nil {
		return
	}
	if length-stat.Size() < 0 {
		err = errors.New("File already exists and is larger than expected size. Aborting.")
		return
	}

	// Now pad the file from the end until it matches required size
	err = fd.Truncate(length)
	if err != nil {
		return
	}

	tfile = &torrentFile{
		path: path,
		lth:  length,
		fd:   fd,
	}

	return
}

func (tf *torrentFile) ReadAt(p []byte, off int64) (n int, err error) {
	n, err = tf.fd.ReadAt(p, off)
	return
}

func (tf *torrentFile) length() int64 {
	return tf.lth
}

func (tf *torrentFile) String() string {
	return fmt.Sprintf("[File: %s Length: %dbytes]", tf.path, tf.length)
}
