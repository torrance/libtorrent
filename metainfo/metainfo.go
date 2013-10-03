package metainfo

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"github.com/zeebo/bencode"
	"io"
	"path/filepath"
)

type Metainfo struct {
	Name         string
	AnnounceList []string
	Pieces       [][]byte
	PieceCount   int
	PieceLength  int64
	InfoHash     []byte
	Files        []struct {
		Length int64
		Path   string
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
		Name:         metaDecode.Info.Name,
		AnnounceList: []string{metaDecode.Announce},
		PieceLength:  metaDecode.Info.PieceLength,
		Pieces:       make([][]byte, len(metaDecode.Info.Pieces)/20),
		PieceCount:   len(metaDecode.Info.Pieces) / 20,
	}

	// Append other announce lists
	for _, list := range metaDecode.List {
		m.AnnounceList = append(m.AnnounceList, list[0])
	}

	// Pieces is a single string of concatenated 20-byte SHA1 hash values for all pieces in the torrent
	// Cycle through and create an slice of hashes
	for i := 0; i < len(metaDecode.Info.Pieces)/20; i++ {
		m.Pieces[i] = metaDecode.Info.Pieces[i*20 : i*20+20]
	}

	// Single files and multiple files are stored differently. We normalise these into
	// a single description
	type file struct {
		Length int64
		Path   string
	}
	if len(metaDecode.Info.Files) == 0 && metaDecode.Info.Length != 0 {
		// Just one file
		m.Files = append(m.Files, file{Length: metaDecode.Info.Length, Path: metaDecode.Info.Name})
	} else {
		// Multiple files
		for _, f := range metaDecode.Info.Files {
			path := filepath.Join(append([]string{metaDecode.Info.Name}, f.Path...)...)
			m.Files = append(m.Files, file{Length: f.Length, Path: path})
		}
	}

	// Create infohash
	h := sha1.New()
	h.Write(metaDecode.RawInfo)
	m.InfoHash = h.Sum(nil)

	return
}
