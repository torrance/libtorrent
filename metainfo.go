package libtorrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"github.com/zeebo/bencode"
	"io"
	"path/filepath"
)

type Metainfo struct {
	name         string
	announceList []string
	pieces       [][]byte
	pieceLength  int64
	infoHash     []byte
	files        []struct {
		length int64
		path   string
	}
}

func ParseMetainfo(r io.Reader) (m *Metainfo, err error) {
	var metaDecode struct {
		Announce string
		List     [][]string         `bencode:"announce-list"`
		RawInfo  bencode.RawMessage `bencode:"info"`
		Info     struct {
			Length      int64
			Name        string
			Pieces      []byte
			PieceLength int64 `bencode:"piece length"`
			Files       []struct {
				Length int64
				Path   []string
			}
		}
	}

	// We need the raw info data to derive the unique info_hash
	// of this torrent. Therefore we decode the metainfo in two steps
	// to obtain both the raw info data and its decoded form.
	dec := bencode.NewDecoder(r)
	if err = dec.Decode(&metaDecode); err != nil {
		return
	}

	dec = bencode.NewDecoder(bytes.NewReader(metaDecode.RawInfo))
	if err = dec.Decode(&metaDecode.Info); err != nil {
		return
	}

	// Basic error checking
	if len(metaDecode.Info.Pieces)%20 != 0 {
		err = errors.New("Metainfo file malformed: Pieces length is not a multiple of 20.")
		return
	}
	// TODO: Other error checking

	// Parse metaDecode into metainfo
	m = &Metainfo{
		name:         metaDecode.Info.Name,
		announceList: []string{metaDecode.Announce},
		pieceLength:  metaDecode.Info.PieceLength,
		pieces:       make([][]byte, len(metaDecode.Info.Pieces)/20),
	}

	// Append other announce lists
	for _, list := range metaDecode.List {
		m.announceList = append(m.announceList, list[0])
	}

	// Pieces is a single string of concatenated 20-byte SHA1 hash values for all pieces in the torrent
	// Cycle through and create an slice of hashes
	for i := 0; i < len(metaDecode.Info.Pieces)/20; i++ {
		m.pieces[i] = metaDecode.Info.Pieces[i*20 : i*20+20]
	}

	// Single files and multiple files are stored differently. We normalise these into
	// a single description
	type file struct {
		length int64
		path   string
	}
	if len(metaDecode.Info.Files) == 0 && metaDecode.Info.Length != 0 {
		// Just one file
		m.files = append(m.files, file{length: metaDecode.Info.Length, path: metaDecode.Info.Name})
	} else {
		// Multiple files
		for _, f := range metaDecode.Info.Files {
			path := filepath.Join(append([]string{metaDecode.Info.Name}, f.Path...)...)
			m.files = append(m.files, file{length: f.Length, path: path})
		}
	}

	// Create infohash
	h := sha1.New()
	h.Write(metaDecode.RawInfo)
	m.infoHash = h.Sum(nil)

	return
}
