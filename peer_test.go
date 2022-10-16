package enet_test

import (
	"github.com/vaeryn-uk/go-enet"
	"testing"
)

func TestPeerData(t *testing.T) {
	const testData uint64 = 5

	// peer is connected to our server.
	// events will produce events as the server receives them.
	peer, events := createServerClient(t)

	// Wait for the server to respond with a connection.
	ev := <-events
	if data, exists := ev.GetPeer().GetDataUint64(); exists {
		t.Fatalf("did not expect new peer to have data set, but has %d", data)
	}

	// Set some data against our peer.
	ev.GetPeer().SetDataUint64(testData)

	// Send a message to the server.
	if err := peer.SendString("testmessage", 0, enet.PacketFlagReliable); err != nil {
		t.Fatal(err)
	}

	// Wait for the server to receive this message, then check the
	// server-side peer associated with this event has the data
	// we set previously.
	ev = <-events
	if data, exists := ev.GetPeer().GetDataUint64(); !exists {
		t.Fatalf("expected peer to have data set, but it wasn't")
	} else if data != testData {
		t.Fatalf("expected peer data to be %d, but it was %d", testData, data)
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
