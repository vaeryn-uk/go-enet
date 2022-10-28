package enet_test

import (
	"github.com/codecat/go-enet"
	"runtime"
	"testing"
)

func TestPeerData(t *testing.T) {
	testData := []byte{0x1, 0x2, 0x3}

	// peer is connected to our server.
	// events will produce events as the server receives them.
	peer, events := createServerClient(t)

	// Wait for the server to respond with a connection.
	ev := <-events
	if data := ev.GetPeer().GetData(); data != nil {
		t.Fatalf("did not expect new peer to have data set, but has %x", data)
	}

	// Set some data against our peer and immediately check it's there.
	ev.GetPeer().SetData(testData)
	assertPeerData(t, ev.GetPeer(), testData, "immediate after set")

	// Send a message to the server.
	if err := peer.SendString("testmessage", 0, enet.PacketFlagReliable); err != nil {
		t.Fatal(err)
	}

	// Wait for the server to receive this message, then check the
	// server-side peer associated with this event has the data
	// we set previously.
	ev = <-events
	assertPeerData(t, ev.GetPeer(), testData, "on packet received")

	// Now do some extra checks on what we can pass in.
	ev.GetPeer().SetData(nil)
	assertPeerData(t, ev.GetPeer(), nil, "nil set")

	// Empty byte slice.
	ev.GetPeer().SetData([]byte{})
	assertPeerData(t, ev.GetPeer(), []byte{}, "empty set")

	// Check that our data stored in C survives garbage collection
	ev.GetPeer().SetData([]byte{1, 2, 3})
	runtime.GC()
	assertPeerData(t, ev.GetPeer(), []byte{1, 2, 3}, "after GC")

	// Maximum length.
	ev.GetPeer().SetData(make([]byte, 0xff))
	assertPeerData(t, ev.GetPeer(), make([]byte, 0xff), "max length")

	// Finally check that anything longer than max panics.
	defer func() {
		if p := recover(); p == nil {
			t.Fatalf("expected SetData() to panic but it didn't")
		}
	}()
	ev.GetPeer().SetData(make([]byte, enet.MaxPeerDataLength+1))
}

func assertPeerData(t testing.TB, peer enet.Peer, expected []byte, msg string) {
	actual := peer.GetData()

	if (actual == nil) != (expected == nil) {
		t.Fatalf("%s: expected peer data to be present? %t vs actual: %t", msg, expected != nil, actual != nil)
	}

	if len(actual) != len(expected) {
		t.Fatalf("%s: expected peer data to have len %d vs actual %d", msg, len(expected), len(actual))
	}

	if string(actual) != string(expected) {
		t.Fatalf("%s: expected peer data to be %v, but it was %v", msg, expected, actual)
	}
}

// createServerClient creates a dummy enet server and client. The returned
// peer can be used to send messages to the server, and the blocking events
// channel returned will be given each event as the server picks it up.
func createServerClient(t *testing.T) (clientConn enet.Peer, serverEvents <-chan enet.Event) {
	port := getFreePort()

	done := make(chan bool, 0)
	events := make(chan enet.Event)

	t.Cleanup(func() {
		// Kill our background service routines for client & server.
		close(done)
	})

	// Create a server and continuously service it, exposing any captured events.
	server, err := enet.NewHost(enet.NewListenAddress(port), 10, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for true {
			select {
			case <-done:
				return
			default:
				ev := server.Service(0)

				// Pass any event out to our channel. This will block
				// until a test consumes it.
				if ev.GetType() != enet.EventNone {
					events <- ev
				}
			}
		}
	}()

	// Create a client and continuously service it in the background.
	client, err := enet.NewHost(nil, 1, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for true {
			select {
			case <-done:
				return
			default:
				client.Service(0)
			}
		}
	}()

	// Connect to our server.
	peer, err := client.Connect(enet.NewAddress("localhost", port), 1, 0)
	if err != nil {
		t.Fatal(err)
	}

	return peer, events
}

var port uint16 = 49152

// getFreePort returns a unique private port. Note this doesn't guarantee
// it's free, but should be good enough from within docker tests.
func getFreePort() uint16 {
	port++
	return port
}
