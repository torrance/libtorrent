package libtorrent

// import (
// 	//"bytes"
// 	//"fmt"
// 	//"io"
// 	"io/ioutil"
// 	//"net"
// 	"os"
// 	"testing"
// )

// func loadTorrentFile(t *testing.T) *os.File {
// 	torrentFile, err := os.Open("testData/test.txt.torrent")
// 	if err != nil {
// 		t.Fatal("Could not open torrent file: ", err)
// 	}
// 	return torrentFile
// }

// func TestNewTorrent(t *testing.T) {
// 	torrentFile := loadTorrentFile(t)

// 	tmpDir, err := ioutil.TempDir("", "libtorrentTesting")
// 	if err != nil {
// 		t.Fatal("Could not create temporary directory to run tests: ", err)
// 	}
// 	defer os.RemoveAll(tmpDir)

// 	config := &Config{RootDirectory: tmpDir}
// 	_, err = NewTorrent(torrentFile, config)
// 	if err != nil {
// 		t.Error("Could not create torrent from valid metainfo file: ", err)
// 	}
// }

//func TestAgainstTransmission(t *testing.T) {
//	torrentFile := loadTorrentFile(t)

//	tmpDir, err := ioutil.TempDir("", "libtorrentTesting")
//	if err != nil {
//		t.Fatal("Could not create temporary directory to run tests: ", err)
//	}
//	defer os.RemoveAll(tmpDir)

//	testFile, _ := os.Create(tmpDir + string(os.PathSeparator) + "test.txt")
//	originalFile, _ := os.Open("testData/test.txt")
//	io.Copy(testFile, originalFile)

//	config := &Config{RootDirectory: tmpDir}
//	tor, err := NewTorrent(torrentFile, config)
//	if err != nil {
//		t.Fatal("Could not create torrent from valid metainfo file: ", err)
//	}

//	if !bytes.Equal(tor.bitf, []byte{0xC0}) {
//		t.Fatalf("Torrent data was not loaded and validated, got: %#x", tor.bitf)
//	}

//	conn, err := net.Dial("tcp", "localhost:51413")
//	if err != nil {
//		t.Fatal("Failed to create connection: ", err)
//	}
//	newHandshake(tor.getInfoHash()).BinaryDump(conn)
//	parseHandshake([][]byte{tor.getInfoHash()}, conn)
//	p := newPeer("transmission", conn)
//	p.write <- &bitfieldMessage{bitf: tor.bitf}
//	tor.swarm = append(tor.swarm, p)
//	fmt.Println("Starting torrent...")
//	tor.start()
//}
