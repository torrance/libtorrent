package libtorrent

import (
	"github.com/torrance/libtorrent/bitfield"
	"io"
	"sync"
	//"testing/iotest"
)

type peer struct {
	name           string
	conn           io.ReadWriter
	write          chan binaryDumper
	read           chan peerDouble
	amChoking      bool
	amInterested   bool
	peerChoking    bool
	peerInterested bool
	mutex          sync.RWMutex
	bitf           *bitfield.Bitfield
}

type peerDouble struct {
	msg  interface{}
	peer *peer
}

func newPeer(name string, conn io.ReadWriter, readChan chan peerDouble) (p *peer) {
	p = &peer{
		name:           name,
		conn:           conn,
		write:          make(chan binaryDumper, 10),
		read:           readChan,
		amChoking:      true,
		amInterested:   false,
		peerChoking:    true,
		peerInterested: false,
	}

	// Write loop
	go func() {
		for {
			//conn := iotest.NewWriteLogger("Writing", conn)
			// TODO: send regular keep alive requests
			msg := <-p.write
			if err := msg.BinaryDump(conn); err != nil {
				// TODO: Close peer
				logger.Error("%s Received error writing to connection: %s", p.name, err)
				return
			}
		}
	}()

	// Read loop
	go func() {
		for {
			//conn := iotest.NewReadLogger("Reading", conn)
			msg, err := parsePeerMessage(conn)

			if _, ok := err.(unknownMessage); ok {
				// Log unknown messages and then ignore
				logger.Info(err.Error())
			} else if err != nil {
				// TODO: Close peer
				logger.Debug("%s Received error reading connection: %s", p.name, err)
				break
			}
			readChan <- peerDouble{msg: msg, peer: p}
		}
	}()

	return
}

func (p *peer) GetAmChoking() (b bool) {
	p.mutex.RLock()
	b = p.amChoking
	p.mutex.RUnlock()
	return
}

func (p *peer) SetAmChoking(b bool) {
	p.mutex.Lock()
	p.amChoking = b
	p.mutex.Unlock()
}

func (p *peer) SetPeerChoking(b bool) {
	p.mutex.Lock()
	p.peerChoking = b
	p.mutex.Unlock()
}

func (p *peer) GetPeerInterested() (b bool) {
	p.mutex.RLock()
	b = p.peerInterested
	p.mutex.RUnlock()
	return
}

func (p *peer) SetPeerInterested(b bool) {
	p.mutex.Lock()
	p.peerInterested = b
	p.mutex.Unlock()
}

func (p *peer) SetBitfield(bitf *bitfield.Bitfield) {
	p.mutex.Lock()
	p.bitf = bitf
	p.mutex.Unlock()
}
