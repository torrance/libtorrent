package tracker

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestTracker(t *testing.T) {
	infoHash := []byte{0x74, 0x2d, 0x47, 0x53, 0x0f, 0xc4, 0xdc, 0xfd, 0xfd, 0x19, 0x71, 0x71, 0xa7, 0x7a, 0x04, 0x88, 0x67, 0xc6, 0xcc, 0x9d}
	stat := &testTorrentStatter{
		infoHash:   infoHash,
		downloaded: 0,
		uploaded:   0,
		left:       36880,
		port:       12345,
	}
	peerChan := make(chan string, 10)
	tkr, _ := NewTracker("udp://tracker.openbittorrent.com:80", stat, peerChan)
	tkr.Start()
	tm := time.After(time.Second * 5)
L:
	for {
		select {
		case <-tm:
			break L
		case p := <-peerChan:
			fmt.Println(p)
		}
	}

	tkr.Stop()
	time.Sleep(time.Second)
}

func TestTrackerPure(t *testing.T) {
	readBuf := bytes.NewBuffer(make([]byte, 16))
	writeBuf := new(bytes.Buffer)
	conn := testConn{
		readBuf:  readBuf,
		writeBuf: writeBuf,
	}
	var udpdialer = func(network, address string) (net.Conn, error) {
		return conn, nil
	}
	UDPDialer = udpdialer

	infoHash := []byte{0x74, 0x2d, 0x47, 0x53, 0x0f, 0xc4, 0xdc, 0xfd, 0xfd, 0x19, 0x71, 0x71, 0xa7, 0x7a, 0x04, 0x88, 0x67, 0xc6, 0xcc, 0x9d}
	stat := &testTorrentStatter{
		infoHash:   infoHash,
		downloaded: 0,
		uploaded:   0,
		left:       36880,
		port:       12345,
	}

	peerChan := make(chan string, 10)
	tkr, _ := NewTracker("udp://tracker.openbittorrent.com:80", stat, peerChan)
	tkr.Start()
	time.Sleep(1000)
	fmt.Println(writeBuf.Bytes())
	fmt.Println(readBuf.Bytes())

}

type testTorrentStatter struct {
	infoHash   []byte
	downloaded int64
	uploaded   int64
	left       int64
	port       int16
}

func (stat *testTorrentStatter) InfoHash() []byte {
	return stat.infoHash
}

func (stat *testTorrentStatter) Downloaded() int64 {
	return stat.downloaded
}

func (stat *testTorrentStatter) Uploaded() int64 {
	return stat.uploaded
}

func (stat *testTorrentStatter) Left() int64 {
	return stat.left
}

func (stat *testTorrentStatter) Port() int16 {
	return stat.port
}

type testConn struct {
	writeBuf *bytes.Buffer
	readBuf  *bytes.Buffer
}

func (conn testConn) Read(b []byte) (n int, err error) {
	n, err = conn.readBuf.Read(b)
	return
}

func (conn testConn) Write(b []byte) (n int, err error) {
	fmt.Println("Trying to write: ", b)
	n, err = conn.writeBuf.Write(b)
	fmt.Println(conn.writeBuf.Bytes())
	return
}

func (conn testConn) Close() (err error) { return }

func (conn testConn) LocalAddr() (addr net.Addr) { return }

func (conn testConn) RemoteAddr() (addr net.Addr) { return }

func (conn testConn) SetDeadline(t time.Time) (err error) { return }

func (conn testConn) SetReadDeadline(t time.Time) (err error) { return }

func (conn testConn) SetWriteDeadline(t time.Time) (err error) { return }
