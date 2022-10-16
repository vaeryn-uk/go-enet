package enet

// #include <enet/enet.h>
import "C"
import (
	"encoding/binary"
	"unsafe"
)

// Peer is a peer which data packets may be sent or received from
type Peer interface {
	GetAddress() Address

	Disconnect(data uint32)
	DisconnectNow(data uint32)
	DisconnectLater(data uint32)

	SendBytes(data []byte, channel uint8, flags PacketFlags) error
	SendString(str string, channel uint8, flags PacketFlags) error
	SendPacket(packet Packet, channel uint8) error

	// SetDataUint64 set an arbitrary values against a peer. This is useful
	// to attach some application-specific data against each peer (such as
	// an identifier).
	//
	// Technically enet allows any data to be stored here (void*), but to keep
	// this type-safe we restrict what can be stored here. New SetDataXXX() methods
	// could be added in the future if needed, e.g. SetDataString() (string, bool).
	//
	// http://enet.bespin.org/structENetPeer.html#a1873959810db7ac7a02da90469ee384e
	SetDataUint64(i uint64)

	// GetDataUint64 returns an application-specific value that's been set
	// against this peer. The bool is true if a value has previously been
	// set.
	//
	// http://enet.bespin.org/structENetPeer.html#a1873959810db7ac7a02da90469ee384e
	GetDataUint64() (uint64, bool)
}

type enetPeer struct {
	cPeer *C.struct__ENetPeer
}

func (peer enetPeer) GetAddress() Address {
	return &enetAddress{
		cAddr: peer.cPeer.address,
	}
}

func (peer enetPeer) Disconnect(data uint32) {
	C.enet_peer_disconnect(
		peer.cPeer,
		(C.enet_uint32)(data),
	)
}

func (peer enetPeer) DisconnectNow(data uint32) {
	C.enet_peer_disconnect_now(
		peer.cPeer,
		(C.enet_uint32)(data),
	)
}

func (peer enetPeer) DisconnectLater(data uint32) {
	C.enet_peer_disconnect_later(
		peer.cPeer,
		(C.enet_uint32)(data),
	)
}

func (peer enetPeer) SendBytes(data []byte, channel uint8, flags PacketFlags) error {
	packet, err := NewPacket(data, flags)
	if err != nil {
		return err
	}
	return peer.SendPacket(packet, channel)
}

func (peer enetPeer) SendString(str string, channel uint8, flags PacketFlags) error {
	packet, err := NewPacket([]byte(str), flags)
	if err != nil {
		return err
	}
	return peer.SendPacket(packet, channel)
}

func (peer enetPeer) SendPacket(packet Packet, channel uint8) error {
	C.enet_peer_send(
		peer.cPeer,
		(C.enet_uint8)(channel),
		packet.(enetPacket).cPacket,
	)
	return nil
}

func (peer enetPeer) SetDataUint64(i uint64) {
	b := make([]byte, 9)
	b[0] = 1
	binary.LittleEndian.PutUint64(b[1:], i)

	peer.cPeer.data = unsafe.Pointer(&b[0])
}

func (peer enetPeer) GetDataUint64() (uint64, bool) {
	if unsafe.Pointer(peer.cPeer.data) == nil {
		return 0, false
	}

	b := C.GoBytes(
		unsafe.Pointer(peer.cPeer.data),
		(C.int)(9),
	)

	if b[0] == 0 {
		return 0, false
	}

	return binary.LittleEndian.Uint64(b[1:]), true
}
