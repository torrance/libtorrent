package tracker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"math/rand"
	"net"
	"net/url"
	"time"
)

const (
	NONE = int32(iota)
	COMPLETED
	STARTED
	STOPPED
)

var logger = logging.MustGetLogger("libtorrent")

// The udpDailer is used to create a udp connection.
// During testing, the udp dialer can be swapped out for a stub.
var UDPDialer func(network, address string) (net.Conn, error) = net.Dial

type TorrentStatter interface {
	InfoHash() []byte
	Downloaded() int64
	Uploaded() int64
	Left() int64
	Port() uint16
	PeerId() []byte
}

type Tracker struct {
	url          *url.URL
	stat         TorrentStatter
	n            uint // This is used like a tcp backoff mechanism
	nextAnnounce time.Duration
	stop         chan struct{}
	peerChan     chan string
	announce     chan struct{} // Used to force an announce
}

type connectRequest struct {
	connection_id  int64
	action         int32
	transaction_id int32
}

func NewTracker(address string, stat TorrentStatter, peerChan chan string) (trk *Tracker, err error) {
	// Verify valid http / or udp address
	url, err := url.Parse(address)
	if err != nil {
		return
	} else if url.Scheme != "udp" {
		err = errors.New(fmt.Sprintf("newTracker: unknown scheme '%s'", url.Scheme))
		return
	}

	trk = &Tracker{
		url:      url,
		stat:     stat,
		peerChan: peerChan,
		stop:     make(chan struct{}),
	}
	return
}

func (tkr *Tracker) Start() {
	tkr.nextAnnounce = 0
	event := STARTED

	go func() {
	L:
		for {
			select {
			case <-time.After(tkr.nextAnnounce):
				// Time to announce
			case <-tkr.announce:
				// We've been forced to announce
			case <-tkr.stop:
				break L
			}

			annReq := &announceRequest{
				transactionId: rand.Int31(),
				infoHash:      tkr.stat.InfoHash(),
				peerId:        tkr.stat.PeerId(),
				downloaded:    tkr.stat.Downloaded(),
				left:          tkr.stat.Left(),
				uploaded:      tkr.stat.Uploaded(),
				port:          tkr.stat.Port(),
				event:         event,
				numWant:       50,
			}
			annRes, err := tkr.udpAnnounce(annReq)
			if err != nil {
				logger.Info("Failed to contact tracker %s, error: %s", tkr.url, err)
				// Attempt again using a backoff pattern 60*2^n
				tkr.nextAnnounce = time.Second * 60 * time.Duration(1<<tkr.n)
				tkr.n++
				continue
			}

			// Success!
			logger.Info("Got %d peers from tracker %s. Next announce in %d seconds", len(annRes.peers), tkr.url, annRes.interval)
			tkr.nextAnnounce = time.Second * time.Duration(annRes.interval)
			event = NONE
			for _, peer := range annRes.peers {
				tkr.peerChan <- peer
			}
		}

		// Announce STOPPED
		annReq := &announceRequest{
			transactionId: rand.Int31(),
			infoHash:      tkr.stat.InfoHash(),
			peerId:        tkr.stat.PeerId(),
			downloaded:    tkr.stat.Downloaded(),
			left:          tkr.stat.Left(),
			uploaded:      tkr.stat.Uploaded(),
			port:          tkr.stat.Port(),
			event:         STOPPED,
			numWant:       50,
		}
		// Ignore failure, we're only making a 'best effort' to shutdown cleanly
		tkr.udpAnnounce(annReq)
	}()
}

func (tkr *Tracker) Stop() {
	close(tkr.stop)
}

func (tkr *Tracker) Announce() {
	go func() { tkr.announce <- struct{}{} }()
}

func (tkr *Tracker) udpAnnounce(annReq *announceRequest) (annRes *announceResponse, err error) {
	conn, err := UDPDialer(tkr.url.Scheme, tkr.url.Host)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(time.Second * 60))

	conReq := &connectionRequest{transactionId: rand.Int31()}
	if err = conReq.BinaryDump(conn); err != nil {
		return
	}

	conRes, err := parseConnectionResponse(conn)
	if err != nil {
		return
	} else if conRes.transactionId != conReq.transactionId {
		err = errors.New("udpAnnounce: recieved transactionId did not match")
		return
	} else if conRes.action != 0 {
		err = errors.New(fmt.Sprintf("udpAnnounce: action is not set to connect (0), instead got %d", conRes.action))
		return
	}

	annReq.connectionId = conRes.connectionId
	if err = annReq.BinaryDump(conn); err != nil {
		return
	}

	if annRes, err = parseAnnounceResponse(conn); err != nil {
		return
	} else if annRes.transactionId != annReq.transactionId {
		errors.New("updAnnounce: received transactionId did not match")
		return
	} else if annRes.action != 1 {
		err = errors.New(fmt.Sprintf("udpAnnounce: action is not set to announce (1), instead got %d", annRes.action))
		return
	}
	return
}

type connectionRequest struct {
	transactionId int32
}

func (c *connectionRequest) BinaryDump(w io.Writer) (err error) {
	// We write out to a buffer first before writing to w, as
	// we need this to go out in a single UDP packet
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, int64(0x41727101980))
	binary.Write(buf, binary.BigEndian, int32(0))
	binary.Write(buf, binary.BigEndian, c.transactionId)
	_, err = w.Write(buf.Bytes())
	return
}

type connectionResponse struct {
	transactionId int32
	connectionId  int64
	action        int32
}

func parseConnectionResponse(r io.Reader) (conRes *connectionResponse, err error) {
	b := make([]byte, 16)
	n, err := r.Read(b)
	if err != nil {
		return
	} else if n < 16 {
		err = errors.New("parseConnectionResponse: UDP packet less than 16 bytes")
		return
	}

	buf := bytes.NewReader(b)
	conRes = new(connectionResponse)
	binary.Read(buf, binary.BigEndian, &conRes.action)
	binary.Read(buf, binary.BigEndian, &conRes.transactionId)
	binary.Read(buf, binary.BigEndian, &conRes.connectionId)
	return
}

type announceRequest struct {
	connectionId  int64
	transactionId int32
	infoHash      []byte
	peerId        []byte
	downloaded    int64
	left          int64
	uploaded      int64
	event         int32
	ipAddress     int32
	key           int32
	numWant       int32
	port          uint16
}

func (a *announceRequest) BinaryDump(w io.Writer) (err error) {
	// Ensure default values are set
	if a.numWant == 0 {
		a.numWant = -1
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, a.connectionId)  // Connection id
	binary.Write(buf, binary.BigEndian, uint32(1))       // Action
	binary.Write(buf, binary.BigEndian, a.transactionId) // Transaction id
	binary.Write(buf, binary.BigEndian, a.infoHash)      // Infohash
	binary.Write(buf, binary.BigEndian, a.peerId)        // Peer id
	binary.Write(buf, binary.BigEndian, a.downloaded)    // Downloaded
	binary.Write(buf, binary.BigEndian, a.left)          // Left
	binary.Write(buf, binary.BigEndian, a.uploaded)      // Uploaded
	binary.Write(buf, binary.BigEndian, a.event)         // Event
	binary.Write(buf, binary.BigEndian, a.ipAddress)     // IP address
	binary.Write(buf, binary.BigEndian, a.key)           // Key
	binary.Write(buf, binary.BigEndian, a.numWant)       // Num want
	binary.Write(buf, binary.BigEndian, a.port)          // Port
	_, err = w.Write(buf.Bytes())
	return
}

type announceResponse struct {
	action        int32
	transactionId int32
	interval      int32
	leechers      int32
	seeders       int32
	peers         []string
}

func parseAnnounceResponse(r io.Reader) (annRes *announceResponse, err error) {
	// Set byte size to equivalent of getting 150 peers
	b := make([]byte, 20+6*150)
	n, err := r.Read(b)
	if err != nil {
		return
	} else if n < 20 {
		err = errors.New("parseAnnounceResponse: response was less than 16 bytes")
		return
	}
	buf := bytes.NewBuffer(b)

	annRes = new(announceResponse)
	binary.Read(buf, binary.BigEndian, &annRes.action)
	binary.Read(buf, binary.BigEndian, &annRes.transactionId)
	binary.Read(buf, binary.BigEndian, &annRes.interval)
	binary.Read(buf, binary.BigEndian, &annRes.leechers)
	binary.Read(buf, binary.BigEndian, &annRes.seeders)

	n = (n - 20) / 6 // Number of ip address + port pairs we read
	for i := 0; i < n; i++ {
		var a, b, c, d uint8
		binary.Read(buf, binary.BigEndian, &a)
		binary.Read(buf, binary.BigEndian, &b)
		binary.Read(buf, binary.BigEndian, &c)
		binary.Read(buf, binary.BigEndian, &d)
		var port uint16
		binary.Read(buf, binary.BigEndian, &port)
		annRes.peers = append(annRes.peers, fmt.Sprintf("%d.%d.%d.%d:%d", a, b, c, d, port))
	}
	return
}
