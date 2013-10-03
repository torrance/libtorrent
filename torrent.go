package libtorrent

import (
	"bytes"
	"fmt"
	"github.com/op/go-logging"
	"github.com/torrance/libtorrent/bitfield"
	"github.com/torrance/libtorrent/filestore"
	"github.com/torrance/libtorrent/tracker"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	Stopped = iota
	Leeching
	Seeding
)

var PeerId = []byte(fmt.Sprintf("libt-%15d", rand.Int63()))[0:20]
var logger = logging.MustGetLogger("libtorrent")

type Torrent struct {
	meta             *Metainfo
	fileStore        *filestore.FileStore
	config           *Config
	bitf             *bitfield.Bitfield
	swarm            []*peer
	incomingPeer     chan *peer
	incomingPeerAddr chan string
	swarmTally       swarmTally
	readChan         chan peerDouble
	trackers         []*tracker.Tracker
	state            int
	stateLock        sync.Mutex
}

func NewTorrent(m *Metainfo, config *Config) (tor *Torrent, err error) {
	tor = &Torrent{
		config:           config,
		meta:             m,
		incomingPeer:     make(chan *peer, 100),
		incomingPeerAddr: make(chan string, 100),
		readChan:         make(chan peerDouble, 50),
		state:            Stopped,
	}

	// Extract file information to create a slice of torrentStorers
	tfiles := make([]filestore.TorrentStorer, 0)
	var tfile filestore.TorrentStorer
	for _, file := range tor.meta.files {
		if tfile, err = filestore.NewTorrentFile(tor.config.RootDirectory, file.path, file.length); err != nil {
			logger.Error("Failed to create file %s: %s", file.path, err)
			return
		}
		tfiles = append(tfiles, tfile)
	}

	// Now we can create our filestore.
	if tor.fileStore, err = filestore.NewFileStore(tfiles, tor.meta.pieces, tor.meta.pieceLength); err != nil {
		logger.Error("Failed to create filestore: %s", err)
		return
	}

	if tor.bitf, err = tor.fileStore.Validate(); err != nil {
		logger.Error("Failed to run validation on new filestore: %s", err)
		return
	}

	return
}

func (tor *Torrent) Start() {
	logger.Info("Torrent starting: %s", tor.meta.name)

	// Set initial state
	tor.stateLock.Lock()
	if tor.bitf.SumTrue() == tor.bitf.Length() {
		tor.state = Seeding
	} else {
		tor.state = Leeching
	}
	tor.stateLock.Unlock()

	// Create trackers
	for _, tkr := range tor.meta.announceList {
		tkr, err := tracker.NewTracker(tkr, tor, tor.incomingPeerAddr)
		if err != nil {
			logger.Error("Failed to create tracker: %s", err)
			continue
		}
		tor.trackers = append(tor.trackers, tkr)
		tkr.Start()
	}

	// Tracker loop
	go func() {
		for {
			peerAddr := <-tor.incomingPeerAddr
			// Only attempt to connect to other peers whilst leeching
			if tor.state != Leeching {
				continue
			}
			go func() {
				conn, err := net.Dial("tcp", peerAddr)
				if err != nil {
					logger.Debug("Failed to connect to tracker peer address %s: %s", peerAddr, err)
					return
				}
				tor.AddPeer(conn, nil)
			}()
		}
	}()

	// Peer loop
	go func() {
		for {
			select {
			case peer := <-tor.incomingPeer:
				// Add to swarm slice
				logger.Debug("Connected to new peer: %s", peer.name)
				tor.swarm = append(tor.swarm, peer)
			case <-time.After(time.Second * 5):
				// Unchoke interested peers
				// TODO: Implement maximum unchoked peers
				// TODO: Implement optimistic unchoking algorithm
				for _, peer := range tor.swarm {
					if peer.GetPeerInterested() && peer.GetAmChoking() {
						logger.Debug("Unchoking peer %s", peer.name)
						peer.write <- &unchokeMessage{}
						peer.SetAmChoking(false)
					}
				}
			}
		}
	}()

	// Receive loop
	go func() {
		for {
			peerDouble := <-tor.readChan
			peer := peerDouble.peer
			msg := peerDouble.msg

			switch msg := msg.(type) {
			case *chokeMessage:
				logger.Debug("Peer %s has choked us", peer.name)
				peer.SetPeerChoking(true)
			case *unchokeMessage:
				logger.Debug("Peer %s has unchoked us", peer.name)
				peer.SetPeerChoking(false)
			case *interestedMessage:
				logger.Debug("Peer %s has said it is interested", peer.name)
				peer.SetPeerInterested(true)
			//case *uninterestedMessage:
			//	logger.Debug("Peer %s has said it is uninterested", peer.name)
			case *haveMessage:
				pieceIndex := int(msg.pieceIndex)
				logger.Debug("Peer %s has piece %d", peer.name, pieceIndex)
				if pieceIndex >= tor.meta.pieceCount {
					logger.Debug("Peer %s sent an out of range have message")
					// TODO: Shutdown client
				}
				peer.HasPiece(pieceIndex)
				// TODO: Update swarmTally
			case *bitfieldMessage:
				logger.Debug("Peer %s has sent us its bitfield", peer.name)
				// Raw parsed bitfield has no actual length. Let's try to set it.
				if err := msg.bitf.SetLength(tor.meta.pieceCount); err != nil {
					logger.Error(err.Error())
					// TODO: Shutdown client
					break
				}
				peer.SetBitfield(msg.bitf)
				tor.swarmTally.AddBitfield(msg.bitf)
			case *requestMessage:
				if peer.GetAmChoking() || !tor.bitf.Get(int(msg.pieceIndex)) || msg.blockLength > 32768 {
					logger.Debug("Peer %s has asked for a block (%d, %d, %d), but we are rejecting them", peer.name, msg.pieceIndex, msg.blockOffset, msg.blockLength)
					// Add naughty points
					break
				}
				logger.Debug("Peer %s has asked for a block (%d, %d, %d), going to fetch block", peer.name, msg.pieceIndex, msg.blockOffset, msg.blockLength)
				block, err := tor.fileStore.GetBlock(int(msg.pieceIndex), int64(msg.blockOffset), int64(msg.blockLength))
				if err != nil {
					logger.Error(err.Error())
					break
				}
				logger.Debug("Peer %s has asked for a block (%d, %d, %d), sending it to them", peer.name, msg.pieceIndex, msg.blockOffset, msg.blockLength)
				peer.write <- &pieceMessage{
					pieceIndex:  msg.pieceIndex,
					blockOffset: msg.blockOffset,
					data:        block,
				}
				// case *pieceMessage:
				// case *cancelMessage:
			default:
				logger.Debug("Peer %s sent unknown message", peer.name)
			}
		}
	}()
}

func (t *Torrent) String() string {
	s := `Torrent: %x
    Name: '%s'
    Piece length: %d
    Announce lists: %v`
	return fmt.Sprintf(s, t.meta.infoHash, t.meta.name, t.meta.pieceLength, t.meta.announceList)
}

func (t *Torrent) InfoHash() []byte {
	return t.meta.infoHash
}

func (t *Torrent) State() (state int) {
	t.stateLock.Lock()
	state = t.state
	t.stateLock.Unlock()
	return
}

func (t *Torrent) AddPeer(conn net.Conn, hs *handshake) {
	// Set 60 second limit to connection attempt
	conn.SetDeadline(time.Now().Add(time.Minute))

	// Send handshake
	if err := newHandshake(t.InfoHash()).BinaryDump(conn); err != nil {
		logger.Debug("%s Failed to send handshake to connection: %s", conn.RemoteAddr(), err)
		return
	}

	// If hs is nil, this means we've attempted to establish the connection and need to wait
	// for their handshake in response
	var err error
	if hs == nil {
		if hs, err = parseHandshake(conn); err != nil {
			logger.Debug("%s Failed to parse incoming handshake: %s", conn.RemoteAddr(), err)
			return
		} else if !bytes.Equal(hs.infoHash, t.InfoHash()) {
			logger.Debug("%s Infohash did not match for connection", conn.RemoteAddr())
			return
		}
	}

	peer := newPeer(string(hs.peerId), conn, t.readChan)
	peer.write <- &bitfieldMessage{bitf: t.bitf}
	t.incomingPeer <- peer

	conn.SetDeadline(time.Time{})
}

func (t *Torrent) Downloaded() int64 {
	// TODO:
	return 0
}

func (t *Torrent) Uploaded() int64 {
	// TODO:
	return 0
}

func (t *Torrent) Left() int64 {
	// TODO:
	return 0
}

func (t *Torrent) Port() int16 {
	return t.config.Port
}

func (t *Torrent) PeerId() []byte {
	return PeerId
}
