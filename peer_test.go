package libtorrent

//import (
//	"fmt"
//	"net"
//	"testing"
//	"testing/iotest"
//	"time"
//)

//func TestPeer(t *testing.T) {
//	conn, err := net.Dial("tcp", "localhost:51413")
//	if err != nil {
//		t.Fatal("Failed to create connection: ", err)
//	}
//	conn.SetDeadline(time.Now().Add(time.Second * 10))

//	logReader := iotest.NewReadLogger("Debugging read", conn)
//	logWriter := iotest.NewWriteLogger("Debugging write", conn)

//	infoHash := []byte{0x74, 0x2d, 0x47, 0x53, 0x0f, 0xc4, 0xdc, 0xfd, 0xfd, 0x19, 0x71, 0x71, 0xa7, 0x7a, 0x04, 0x88, 0x67, 0xc6, 0xcc, 0x9d}

//	hs := newHandshake(infoHash)
//	err = hs.BinaryDump(logWriter)
//	if err != nil {
//		t.Fatal("Failed to write handshake to connection")
//	}

//	hs, err = parseHandshake([][]byte{infoHash}, logReader)
//	if err != nil {
//		t.Fatal("Failed to parse handshake: ", err)
//	}
//	fmt.Println(hs)

//	msg, err := parsePeerMessage(logReader)
//	if err != nil {
//		t.Fatal("Failed to parse peer message [1]: ", err)
//	}
//	fmt.Println("Message [1]: ", msg)

//	// Send have
//	fmt.Println("Send bitfield")
//	msg1 := &bitfieldMessage{
//		bitf: newBitfield(2),
//	}
//	msg1.BinaryDump(logWriter)
//	fmt.Println("Finished bitfield")

//	// Say interested
//	fmt.Println("Saying interested")
//	msg2 := &interestedMessage{}
//	err = msg2.BinaryDump(logWriter)
//	if err != nil {
//		t.Error("Failed to write interested msg: ", err)
//	}
//	fmt.Println("Finished interested")

//	msg, err = parsePeerMessage(logReader)
//	if err != nil {
//		t.Fatal("Failed to parse peer message [2]: ", err)
//	}
//	fmt.Println("Message [2]: ", msg)
//}
